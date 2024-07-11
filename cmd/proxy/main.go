package main

import (
	"desc/services/youtube"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

func main() {

	rawProxies := youtube.GetProxyList("cmd/proxy/files/raw_proxies.txt")

	filteredProxies := []string{}
	for _, proxy := range rawProxies {
		proxy = strings.ReplaceAll(proxy, "\r", "")
		proxy := fmt.Sprintf("http://%s", proxy)
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Fatalln(err)
		}
		content, err := youtube.RequestVideoHTML("KkCXLABwHP0", proxyUrl)

		if err != nil {
			fmt.Println(fmt.Sprintf("FAIL: %s", proxy))
			continue
		}
		if content == "" {
			fmt.Println(fmt.Sprintf("NO ERROR: %s", proxy))
			continue
		}
		fmt.Println(fmt.Sprintf("Success: %s", proxy))
		filteredProxies = append(filteredProxies, proxy)
	}

	final := strings.Join(filteredProxies, "\n")
	WriteTextToFile(final, "cmd/proxy/files/filtered_proxies.txt")
}

func WriteTextToFile(text string, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Write the text to the file
	err := os.WriteFile(filePath, []byte(text), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
