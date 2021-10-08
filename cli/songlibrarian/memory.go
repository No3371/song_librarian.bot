package main

var redirectedUrlMap map[string]struct{}
var redirectedUrls []string

func init () {
	redirectedUrlMap = make(map[string]struct{}, 32)
	redirectedUrls = make([]string, 0)
}

func isRedirected (url string) {
	
}

func markRedirected (url string) {
	if _, exists := redirectedUrlMap[url]; exists {
		return
	}

	redirectedUrlMap[url] = struct{}{}
	redirectedUrls = append(redirectedUrls, url)

}