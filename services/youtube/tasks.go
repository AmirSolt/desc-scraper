package youtube

import (
	"context"
	"desc/base"
	"desc/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
Future Fixes:
1. convertVideoHTMLToObject is a giant poop
*/

const NumTasks int = 24

func RunTasks(b *base.Base) {

	// b.MemQ.Enqueue("3hw_y9hI_js")

	var wg sync.WaitGroup
	wg.Add(NumTasks)
	for i := 0; i < NumTasks; i++ {
		taskerName := fmt.Sprintf("T%d", i)
		go func() {
			defer wg.Done()
			fmt.Println(fmt.Sprintf("%s: Started Running", taskerName))
			if err := VideoScrapeTask(b, taskerName); err != nil {
				fmt.Println(fmt.Sprintf("%s: Error running task:", taskerName), err)
			}
		}()
	}
	wg.Wait()
}

func VideoScrapeTask(b *base.Base, taskerName string) error {
	ctx := context.Background()

	size, err := b.MemQ.Size()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(fmt.Sprintf("%s - Queue Size: %d", taskerName, size))

	t1 := time.Now()
	totalReq := 0

	for true {
		vidID, err := b.MemQ.Dequeue()
		if err != nil {
			log.Fatal(err)
			return err
		}
		if vidID == "" {
			err := fmt.Errorf("%s - Error: video_queue is empty", taskerName)
			log.Fatal(err)
			return err
		}

		vidHTML, err := requestVideoHTML(vidID)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s - WARNING: Request Failed: %s", taskerName, err.Error()))
			continue
		}

		videoResult, err2 := convertVideoHTMLToObject(vidHTML)
		if err2 != nil {
			fmt.Println(fmt.Sprintf("%s - WARNING: Convertion Failed: %s", taskerName, err2.Error()))
			continue
		}

		queueSize, err := b.MemQ.Size()
		if err != nil {
			log.Fatal(err)
			return err
		}
		if queueSize < b.Config.MaxQueueSize {
			var vidIDs []string
			for _, compactVid := range videoResult.compactVideoRenderers {
				if compactVid.VideoID == "" {
					continue
				}
				vidIDs = append(vidIDs, compactVid.VideoID)
			}
			b.MemQ.EnqueueAll(vidIDs)
		}

		channel, err := findSertChannel(b, ctx, videoResult)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s - WARNING: Channel Finsert Failed: %s", taskerName, err.Error()))
			continue
		}
		_, err = findSertVideo(b, ctx, channel, videoResult, vidID)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s - WARNING: Video Finsert Failed: %s", taskerName, err.Error()))
			continue
		}

		totalReq++
		if totalReq%10 == 0 {
			elapsed := time.Now().Sub(t1).Seconds()
			if elapsed > 0 {
				reqRate := float64(totalReq) / elapsed
				fmt.Println(fmt.Sprintf("%s - Queue Size: %d", taskerName, queueSize))
				fmt.Println(fmt.Sprintf("%s - Request Rate (req/s): %f", taskerName, reqRate))
			}
		}
		// time.Sleep(10 * time.Millisecond)
		// fmt.Println(fmt.Sprintf(">>> Loop Count: %d", count))
	}

	fmt.Println(b.MemQ.Size())
	return nil
}

func findSertChannel(b *base.Base, ctx context.Context, videoResult *VideoResult) (*models.Channel, error) {
	ytID := videoResult.videoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Title.Runs[0].NavigationEndpoint.BrowseEndpoint.BrowseId
	channel, err := b.DB.Queries.GetChannelByYTID(ctx, ytID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if channel.YtID != "" {
		return &channel, nil
	}

	thumbnails := videoResult.videoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Thumbnail.Thumbnails
	channel, err = b.DB.Queries.CreateChannel(ctx, models.CreateChannelParams{
		YtID:         ytID,
		ThumbnailUrl: thumbnails[len(thumbnails)-1].URL,
		Handle:       videoResult.videoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Title.Runs[0].NavigationEndpoint.BrowseEndpoint.CanonicalBaseUrl,
		Title:        videoResult.videoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Title.Runs[0].Text,
	})
	if err != nil {
		return nil, err
	}
	return &channel, nil
}
func findSertVideo(b *base.Base, ctx context.Context, channel *models.Channel, videoResult *VideoResult, vidYTID string) (*models.Video, error) {
	video, err := b.DB.Queries.GetVideoByYTID(ctx, vidYTID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return &video, nil
	}

	params := models.CreateVideoParams{
		YtID:        vidYTID,
		Title:       videoResult.videoPrimaryInfoRenderer.Title.Runs[0].Text,
		Description: videoResult.videoSecondaryInfoRenderer.AttributedDescription.Content,
		ChannelID:   channel.ID,
	}

	date, err := dateparse.ParseAny(videoResult.videoPrimaryInfoRenderer.DateText.SimpleText)
	if err == nil {
		params.PublishedAt = pgtype.Timestamptz{Time: date}
	}

	video, err = b.DB.Queries.CreateVideo(ctx, params)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func convertVideoHTMLToObject(vidHTML string) (*VideoResult, error) {
	jsonStr, err := extractTextBetweenMarkers(vidHTML)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}

	// Helper function to safely get a value from a map and assert its type
	getMap := func(m map[string]interface{}, key string) (map[string]interface{}, error) {
		if val, ok := m[key]; ok {
			if casted, ok := val.(map[string]interface{}); ok {
				return casted, nil
			}
		}
		return nil, fmt.Errorf("key %s does not contain a map[string]interface{}", key)
	}

	getSlice := func(m map[string]interface{}, key string) ([]interface{}, error) {
		if val, ok := m[key]; ok {
			if casted, ok := val.([]interface{}); ok {
				return casted, nil
			}
		}
		return nil, fmt.Errorf("key %s does not contain a []interface{}", key)
	}

	getElementMap := func(slice []interface{}, index int) (map[string]interface{}, error) {
		if index < len(slice) {
			if casted, ok := slice[index].(map[string]interface{}); ok {
				return casted, nil
			}
		}
		return nil, fmt.Errorf("index %d does not contain a map[string]interface{}", index)
	}

	contents, err := getMap(result, "contents")
	if err != nil {
		return nil, err
	}

	twoColumnWatchNextResults, err := getMap(contents, "twoColumnWatchNextResults")
	if err != nil {
		return nil, err
	}

	results, err := getMap(twoColumnWatchNextResults, "results")
	if err != nil {
		return nil, err
	}

	resultContents, err := getMap(results, "results")
	if err != nil {
		return nil, err
	}

	contents2, err := getSlice(resultContents, "contents")
	if err != nil {
		return nil, err
	}

	videoPrimaryInfoRendererElem, err := getElementMap(contents2, 0)
	if err != nil {
		return nil, err
	}

	videoPrimaryInfoRendererMap, err := getMap(videoPrimaryInfoRendererElem, "videoPrimaryInfoRenderer")
	if err != nil {
		return nil, err
	}

	var videoPrimaryInfoRenderer VideoPrimaryInfoRenderer
	if err := convertMapToStruct(videoPrimaryInfoRendererMap, &videoPrimaryInfoRenderer); err != nil {
		return nil, err
	}

	videoSecondaryInfoRendererElem, err := getElementMap(contents2, 1)
	if err != nil {
		return nil, err
	}

	videoSecondaryInfoRendererMap, err := getMap(videoSecondaryInfoRendererElem, "videoSecondaryInfoRenderer")
	if err != nil {
		return nil, err
	}

	var videoSecondaryInfoRenderer VideoSecondaryInfoRenderer
	if err := convertMapToStruct(videoSecondaryInfoRendererMap, &videoSecondaryInfoRenderer); err != nil {
		return nil, err
	}

	secondaryResults, err := getMap(twoColumnWatchNextResults, "secondaryResults")
	if err != nil {
		return nil, err
	}

	secondaryResults2, err := getMap(secondaryResults, "secondaryResults")
	if err != nil {
		return nil, err
	}

	secondaryResultsResults, err := getSlice(secondaryResults2, "results")
	if err != nil {
		return nil, err
	}

	var compactVideoRenderers []CompactVideoRenderer
	for _, secResult := range secondaryResultsResults {
		if resultMap, ok := secResult.(map[string]interface{}); ok {
			if cvRendererMap, ok := resultMap["compactVideoRenderer"].(map[string]interface{}); ok {
				var compactVideoRenderer CompactVideoRenderer
				if err := convertMapToStruct(cvRendererMap, &compactVideoRenderer); err != nil {
					return nil, err
				}
				compactVideoRenderers = append(compactVideoRenderers, compactVideoRenderer)
			}
		}
	}

	return &VideoResult{
		videoPrimaryInfoRenderer:   videoPrimaryInfoRenderer,
		videoSecondaryInfoRenderer: videoSecondaryInfoRenderer,
		compactVideoRenderers:      compactVideoRenderers,
	}, nil
}

// Helper function to convert a map to a struct using JSON marshalling/unmarshalling
func convertMapToStruct(m map[string]interface{}, v interface{}) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

func extractTextBetweenMarkers(text string) (string, error) {
	// Define the regex pattern
	pattern := `ytInitialData\s*=\s*(.*?)\s*;\s*<\/script>`

	// Compile the regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	// Find the substring that matches the pattern
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", fmt.Errorf("no match found")
	}

	// Return the matched group
	return matches[1], nil
}

func requestVideoHTML(vidID string) (string, error) {
	return getYtRequest(fmt.Sprintf("https://www.youtube.com/watch?v=%s", vidID))
}

func getYtRequest(url string) (string, error) {
	// Create a new HTTP client
	client := getSmartProxyClient()

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request: %v", err)
	}
	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the HTTP status code is OK
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		time.Sleep(5 * time.Second)
		return "", fmt.Errorf("delayed status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func getRandom(min, max int) int {
	return rand.Intn(max-min) + min
}

func getSmartProxyClient() *http.Client {
	proxyUrl, err := url.Parse(fmt.Sprintf("http://sp3z4sznsk:h+lkSNLmL5f9gto02c@dc.smartproxy.com:%d", randomNumber(10001, 10100)))
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}
	return client
}

func randomNumber(min, max int) int {
	// rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
