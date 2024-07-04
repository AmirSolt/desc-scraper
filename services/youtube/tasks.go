package youtube

import (
	"desc/base"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

func RunTasks(b *base.Base) error {
	return VideoScrapeTask(b)
}

func VideoScrapeTask(b *base.Base) error {

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
			log.Fatal(err)
			return err
		}

		videoResult, err2 := convertVideoHTMLToObject(vidHTML)
		if err2 != nil {
			log.Fatal(err2)
			return err2
		}

		queueSize, err := b.MemQ.Size()
		if err != nil {
			log.Fatal(err)
			return err
		}
		if queueSize < b.Config.MaxQueueSize {
			for _, compactVid := range videoResult.compactVideoRenderers {
				b.MemQ.Enqueue(compactVid.VideoID)
			}
		}

		// add to DB
		// fmt.Println(videoResult)

		time.Sleep(1 * time.Second)

	}

	return nil
}

func convertVideoHTMLToObject(vidHTML string) (*VideoResult, error) {

	jsonStr, err := extractTextBetweenMarkers(vidHTML)
	if err != nil {
		return nil, err
	}

	os.WriteFile("raw.json", []byte(jsonStr), os.ModePerm)

	var result map[string]interface{}
	errUnm := json.Unmarshal([]byte(jsonStr), &result)
	if errUnm != nil {
		return nil, errUnm
	}

	// =============================================
	// Navigate to the videoPrimaryInfoRenderer part of the JSON
	contents := result["contents"].(map[string]interface{})
	twoColumnWatchNextResults := contents["twoColumnWatchNextResults"].(map[string]interface{})
	results := twoColumnWatchNextResults["results"].(map[string]interface{})
	resultContents := results["results"].(map[string]interface{})
	contents2 := resultContents["contents"].([]interface{})
	videoPrimaryInfoRendererMap := contents2[0].(map[string]interface{})["videoPrimaryInfoRenderer"].(map[string]interface{})

	// Convert the segment to JSON
	videoPrimaryInfoRendererJSON, errMar := json.Marshal(videoPrimaryInfoRendererMap)
	if errMar != nil {
		return nil, errMar
	}

	// Unmarshal the JSON to the struct
	var videoPrimaryInfoRenderer VideoPrimaryInfoRenderer
	errUnm2 := json.Unmarshal(videoPrimaryInfoRendererJSON, &videoPrimaryInfoRenderer)
	if errUnm2 != nil {
		return nil, errUnm2
	}
	// =============================================

	// =============================================
	// Navigate to the videoPrimaryInfoRenderer part of the JSON
	videoSecondaryInfoRendererMap := contents2[1].(map[string]interface{})["videoSecondaryInfoRenderer"].(map[string]interface{})

	// Convert the segment to JSON
	videoSecondaryInfoRendererMapJSON, errMar := json.Marshal(videoSecondaryInfoRendererMap)
	if errMar != nil {
		return nil, errMar
	}

	// Unmarshal the JSON to the struct
	var videoSecondaryInfoRenderer VideoSecondaryInfoRenderer
	errUnm3 := json.Unmarshal(videoSecondaryInfoRendererMapJSON, &videoSecondaryInfoRenderer)
	if errUnm3 != nil {
		return nil, errUnm3
	}
	// =============================================

	// =============================================
	// Navigate to the videoPrimaryInfoRenderer part of the JSON
	secondaryResults := twoColumnWatchNextResults["secondaryResults"].(map[string]interface{})
	secondaryResults2 := secondaryResults["secondaryResults"].(map[string]interface{})
	secondaryResultsResults := secondaryResults2["results"].([]interface{})

	var compactVideoRenderers []CompactVideoRenderer
	for _, secondaryResultsResult := range secondaryResultsResults {
		srr := secondaryResultsResult.(map[string]interface{})
		if _, ok := srr["compactVideoRenderer"]; ok {
			compactVideoRendererMap := srr["compactVideoRenderer"].(map[string]interface{})
			// Convert the segment to JSON
			compactVideoRendererMapJson, err := json.Marshal(compactVideoRendererMap)
			if err != nil {
				return nil, err
			}

			// Unmarshal the JSON to the struct
			var compactVideoRenderer CompactVideoRenderer
			errUnm3 := json.Unmarshal(compactVideoRendererMapJson, &compactVideoRenderer)
			if errUnm3 != nil {
				return nil, errUnm3
			}

			compactVideoRenderers = append(compactVideoRenderers, compactVideoRenderer)
		}
	}
	// =============================================

	// "twoColumnWatchNextResults"
	// "results"
	// "results"
	// "contents"[0]"videoPrimaryInfoRenderer"

	// "secondaryResults"
	// "results"[x] if "compactVideoRenderer"

	return &VideoResult{
		videoPrimaryInfoRenderer:   videoPrimaryInfoRenderer,
		videoSecondaryInfoRenderer: videoSecondaryInfoRenderer,
		compactVideoRenderers:      compactVideoRenderers,
	}, nil
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
