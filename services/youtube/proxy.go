package youtube

import (
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
)

// func getSmartProxyClient() *http.Client {
// 	proxy := fmt.Sprintf("http://sp3z4sznsk:h+lkSNLmL5f9gto02c@dc.smartproxy.com:%d", randomNumber(10001, 10100))
// 	proxyUrl, err := url.Parse(proxy)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	client := &http.Client{
// 		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
// 	}
// 	return client
// }

// func getRandom(min, max int) int {
// 	return rand.Intn(max-min) + min
// }

// func randomNumber(min, max int) int {
// 	// rand.Seed(time.Now().UnixNano())
// 	return rand.Intn(max-min) + min
// }

// ====================================

func GetProxyList(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	b, err := io.ReadAll(file)
	return strings.Split(string(b), "\n")
}

func getRandomProxyURL(proxies []string) *url.URL {
	proxy := proxies[rand.Intn(len(proxies))]
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		log.Fatalln(err)
	}
	return proxyUrl
}
