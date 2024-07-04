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
	"net/http"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
Future Fixes:

1. convertVideoHTMLToObject is a giant poop

*/

func RunTasks(b *base.Base) error {
	return VideoScrapeTask(b)
}

func VideoScrapeTask(b *base.Base) error {
	ctx := context.Background()

	size, err := b.MemQ.Size()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(fmt.Sprintf("Queue Size: %d", size))

	count := 0
	for true {
		vidID, err := b.MemQ.Dequeue()
		if err != nil {
			log.Fatal(err)
			return err
		}
		if vidID == "" {
			err := fmt.Errorf("video_queue is empty")
			log.Fatal(err)
			return err
		}

		vidHTML, err := requestVideoHTML(vidID)
		if err != nil {
			continue
		}

		videoResult, err2 := convertVideoHTMLToObject(vidHTML)
		if err2 != nil {
			continue
		}

		queueSize, err := b.MemQ.Size()
		if err != nil {
			log.Fatal(err)
			return err
		}
		fmt.Println(fmt.Sprintf("Queue Size: %d", queueSize))
		if queueSize < b.Config.MaxQueueSize {
			var vidIDs []string
			for _, compactVid := range videoResult.compactVideoRenderers {
				vidIDs = append(vidIDs, compactVid.VideoID)
			}
			b.MemQ.EnqueueAll(vidIDs)
		}

		channel, err := findSertChannel(b, ctx, videoResult)
		if err != nil {
			log.Fatal(err)
			return err
		}
		_, err = findSertVideo(b, ctx, channel, videoResult, vidID)
		if err != nil {
			log.Fatal(err)
			return err
		}

		count++
		time.Sleep(1 * time.Second)
		fmt.Println(fmt.Sprintf(">>> Loop Count: %d", count))
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
	if err == nil {
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

// func convertVideoHTMLToObject(vidHTML string) (*VideoResult, error) {

// 	jsonStr, err := extractTextBetweenMarkers(vidHTML)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result map[string]interface{}
// 	errUnm := json.Unmarshal([]byte(jsonStr), &result)
// 	if errUnm != nil {
// 		return nil, errUnm
// 	}

// 	// =============================================
// 	// Navigate to the videoPrimaryInfoRenderer part of the JSON
// 	contents := result["contents"].(map[string]interface{})
// 	twoColumnWatchNextResults := contents["twoColumnWatchNextResults"].(map[string]interface{})
// 	results := twoColumnWatchNextResults["results"].(map[string]interface{})
// 	resultContents := results["results"].(map[string]interface{})
// 	contents2 := resultContents["contents"].([]interface{})
// 	videoPrimaryInfoRendererMap := contents2[0].(map[string]interface{})["videoPrimaryInfoRenderer"].(map[string]interface{})

// 	// Convert the segment to JSON
// 	videoPrimaryInfoRendererJSON, errMar := json.Marshal(videoPrimaryInfoRendererMap)
// 	if errMar != nil {
// 		return nil, errMar
// 	}

// 	// Unmarshal the JSON to the struct
// 	var videoPrimaryInfoRenderer VideoPrimaryInfoRenderer
// 	errUnm2 := json.Unmarshal(videoPrimaryInfoRendererJSON, &videoPrimaryInfoRenderer)
// 	if errUnm2 != nil {
// 		return nil, errUnm2
// 	}
// 	// =============================================

// 	// =============================================
// 	// Navigate to the videoPrimaryInfoRenderer part of the JSON
// 	videoSecondaryInfoRendererMap := contents2[1].(map[string]interface{})["videoSecondaryInfoRenderer"].(map[string]interface{})

// 	// Convert the segment to JSON
// 	videoSecondaryInfoRendererMapJSON, errMar := json.Marshal(videoSecondaryInfoRendererMap)
// 	if errMar != nil {
// 		return nil, errMar
// 	}

// 	// Unmarshal the JSON to the struct
// 	var videoSecondaryInfoRenderer VideoSecondaryInfoRenderer
// 	errUnm3 := json.Unmarshal(videoSecondaryInfoRendererMapJSON, &videoSecondaryInfoRenderer)
// 	if errUnm3 != nil {
// 		return nil, errUnm3
// 	}
// 	// =============================================

// 	// =============================================
// 	// Navigate to the videoPrimaryInfoRenderer part of the JSON
// 	secondaryResults := twoColumnWatchNextResults["secondaryResults"].(map[string]interface{})
// 	secondaryResults2 := secondaryResults["secondaryResults"].(map[string]interface{})
// 	secondaryResultsResults := secondaryResults2["results"].([]interface{})

// 	var compactVideoRenderers []CompactVideoRenderer
// 	for _, secondaryResultsResult := range secondaryResultsResults {
// 		srr := secondaryResultsResult.(map[string]interface{})
// 		if _, ok := srr["compactVideoRenderer"]; ok {
// 			compactVideoRendererMap := srr["compactVideoRenderer"].(map[string]interface{})
// 			// Convert the segment to JSON
// 			compactVideoRendererMapJson, err := json.Marshal(compactVideoRendererMap)
// 			if err != nil {
// 				return nil, err
// 			}

// 			// Unmarshal the JSON to the struct
// 			var compactVideoRenderer CompactVideoRenderer
// 			errUnm3 := json.Unmarshal(compactVideoRendererMapJson, &compactVideoRenderer)
// 			if errUnm3 != nil {
// 				return nil, errUnm3
// 			}

// 			compactVideoRenderers = append(compactVideoRenderers, compactVideoRenderer)
// 		}
// 	}
// 	// =============================================

// 	// "twoColumnWatchNextResults"
// 	// "results"
// 	// "results"
// 	// "contents"[0]"videoPrimaryInfoRenderer"

// 	// "secondaryResults"
// 	// "results"[x] if "compactVideoRenderer"

// 	return &VideoResult{
// 		videoPrimaryInfoRenderer:   videoPrimaryInfoRenderer,
// 		videoSecondaryInfoRenderer: videoSecondaryInfoRenderer,
// 		compactVideoRenderers:      compactVideoRenderers,
// 	}, nil
// }

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
	client := &http.Client{}

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
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}
