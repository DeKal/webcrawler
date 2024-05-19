package main

import "github.com/DeKal/webcrawler"

func main() {
	links := []string{"https://www.youtube.com/"}
	webcrawler := webcrawler.WebCrawler{
		Mode:           webcrawler.Async,
		MaxConcurrency: 32,
	}
	webcrawler.FindAndDisPlay(links, 3)

}
