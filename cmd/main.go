package main

import "github.com/DeKal/webcrawler"

func main() {
	links := []string{"https://simplewebserver.org/"}
	webcrawler := webcrawler.WebCrawler{Mode: webcrawler.Sync}
	webcrawler.FindAndDisPlay(links, 1)

}
