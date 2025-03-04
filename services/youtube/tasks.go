package youtube

import (
	"context"
	"desc/base"
	"desc/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

/*
Future Fixes:
1. convertVideoHTMLToObject is a giant poop
*/

var proxyIndex int = 0
var totalReq int = 0
var t1 time.Time = time.Now()

func RunTasks(b *base.Base) {

	queue := &Queue{queue: []string{"KkCXLABwHP0"}}

	var wg sync.WaitGroup
	wg.Add(b.Env.NUMBER_OF_INSTANCES)
	for i := 0; i < b.Env.NUMBER_OF_INSTANCES; i++ {
		taskerName := fmt.Sprintf("T%d", i)
		go func() {
			defer wg.Done()
			fmt.Println(fmt.Sprintf("%s: Started Running", taskerName))
			if err := VideoScrapeTask(b, taskerName, queue); err != nil {
				fmt.Println(fmt.Sprintf("%s: Error running task:", taskerName), err)
			}
		}()
	}
	wg.Wait()
}

func VideoScrapeTask(b *base.Base, taskerName string, queue *Queue) error {
	ctx := context.Background()
	proxies := GetProxyList("cmd/proxy/files/filtered_proxies.txt")

	fmt.Println(fmt.Sprintf("%s - Queue Size: %d", taskerName, queue.Size()))

	for true {
		vidID, ok := queue.Dequeue()
		if !ok {
			// err := fmt.Errorf("%s - Error: video_queue is empty", taskerName)
			// log.Fatal(err)
			// return err
			fmt.Println(fmt.Sprintf("%s - WARNING: video_queue is empty", taskerName))
			time.Sleep(10 * time.Second)
			continue
		}

		proxy := getProxyURL(proxies[proxyIndex])
		proxyIndex = (proxyIndex + 1) % len(proxies)
		vidHTML, err := RequestVideoHTML(vidID, proxy)
		if err != nil {
			// log.Fatal(err)
			// return err
			fmt.Println(fmt.Sprintf("%s - WARNING: Request Failed: %s", taskerName, err.Error()))
			queue.Enqueue(vidID)
			continue
		}
		if vidHTML == "" {
			fmt.Println(fmt.Sprintf("%s - WARNING: Request Unsuccessful: %s", taskerName, err.Error()))
			queue.Enqueue(vidID)
			continue
		}

		videoResult, err2 := convertVideoHTMLToObject(vidHTML)
		if err2 != nil {
			fmt.Println(fmt.Sprintf("%s - WARNING: Convertion Failed: %s", taskerName, err2.Error()))
			continue
		}

		if queue.Size() < b.Config.MaxQueueSize {
			var vidIDs []string
			for _, compactVid := range videoResult.compactVideoRenderers {
				if compactVid.VideoID == "" {
					continue
				}
				vidIDs = append(vidIDs, compactVid.VideoID)
			}
			queue.EnqueueAll(vidIDs)
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
		if totalReq%100 == 0 {
			elapsed := time.Since(t1).Seconds()
			if elapsed > 0 {
				reqRate := float64(totalReq) / elapsed
				fmt.Println(fmt.Sprintf("Request Rate (req/s): %f - Queue Size: %d", reqRate, queue.Size()))
			}
		}
	}

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
	if len(videoResult.videoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Title.Runs) <= 0 {
		return nil, fmt.Errorf("videoResult.videoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Title.Runs is empty")
	}
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

	if len(videoResult.videoPrimaryInfoRenderer.Title.Runs) <= 0 {
		return nil, fmt.Errorf("videoResult.videoPrimaryInfoRenderer.Title.Runs is empty")
	}

	params := models.CreateVideoParams{
		YtID:        vidYTID,
		Title:       videoResult.videoPrimaryInfoRenderer.Title.Runs[0].Text,
		Description: videoResult.videoSecondaryInfoRenderer.AttributedDescription.Content,
		ChannelID:   channel.ID,
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

func RequestVideoHTML(vidID string, proxy *url.URL) (string, error) {
	return getYtRequest(fmt.Sprintf("https://www.youtube.com/watch?v=%s", vidID), proxy)
}

func getYtRequest(url string, proxy *url.URL) (string, error) {
	// Create a new HTTP client

	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxy)},
		// Timeout:   120 * time.Second,
	}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("proxy: %s - failed to create GET request: %v", proxy.Host, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	// Send the request
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("proxy: %s - failed to send GET request: %v", proxy.Host, err)
	}
	defer resp.Body.Close()

	// Check if the HTTP status code is OK
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
		return "", fmt.Errorf("proxy: %s - unexpected status code: %d", proxy.Host, resp.StatusCode)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		time.Sleep(5 * time.Second)
		return "", nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("proxy: %s - failed to read response body: %v", proxy.Host, err)
	}

	return string(body), nil
}
