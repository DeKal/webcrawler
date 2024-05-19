package webcrawler

import (
	"fmt"
	"sync"

	"github.com/DeKal/webcrawler/internal/crawlutils"
	"github.com/DeKal/webcrawler/internal/httpParser"
)

type Mode int64

const (
	Sync Mode = iota
	Async
)

type WebCrawler struct {
	Mode Mode
}

func (w WebCrawler) getCrawlFunc() func(rootLink string, depth int) map[string][]string {
	switch w.Mode {
	case Sync:
		return w.crawl
	case Async:
		return w.asyncCrawlWithChannel
	default:
		return w.crawl
	}
}

func (w WebCrawler) Find(rootLinks []string, depth int) map[string][]string {
	crawl := w.getCrawlFunc()
	allLinks := make(map[string][]string)
	for _, rootLink := range rootLinks {
		allLinks = crawlutils.CombineMaps(allLinks, crawl(rootLink, depth))
	}
	return allLinks
}

func (w WebCrawler) FindAndDisPlay(rootLinks []string, depth int) {
	w.display(w.Find(rootLinks, depth))
}

func (WebCrawler) display(linksMap map[string][]string) {
	fmt.Println("Here is the result:")
	for link, subLinks := range linksMap {
		fmt.Println("From Link ", link)
		for _, subLink := range subLinks {
			fmt.Println("-------> Sub Link ", subLink)
		}
	}
}

func (WebCrawler) crawl(rootLink string, depth int) map[string][]string {
	var (
		allLinksMap = make(map[string][]string)
		q           = make([]struct {
			string
			int
		}, 0)
		visited = make(map[string]bool)
	)

	q = append(q, struct {
		string
		int
	}{rootLink, 0})

	for len(q) > 0 {
		top := q[0]
		topLink := top.string
		topLevel := top.int
		if topLevel <= depth {
			fmt.Println("exec go routine for link ", topLink, " and level ", topLevel)
			allLinksMap[topLink] = httpParser.GetUniqueLinksFromUrl(topLink)
			for _, link := range allLinksMap[topLink] {
				if _, ok := visited[link]; !ok {
					q = append(q, struct {
						string
						int
					}{link, topLevel + 1})
					visited[link] = true
				}
			}
			fmt.Println("done go routine for link ", topLink, " and level ", topLevel)
		}

		q = q[1:]
	}

	return allLinksMap
}

func (WebCrawler) asyncCrawlWithChannel(rootLink string, depth int) map[string][]string {
	var (
		allLinksMap = make(map[string][]string)
		q           = make(chan struct {
			string
			int
		}, 5000)
		size    int
		visited = make(map[string]bool)
		wg      sync.WaitGroup
	)

	q <- struct {
		string
		int
	}{rootLink, 0}
	size = 1

	for size > 0 {
		top := <-q
		size--

		wg.Add(1)

		go func(topLink string, topLevel int, size *int) {
			fmt.Println("exec go routine for link ", topLink, " and level ", topLevel)
			defer wg.Done()
			if topLevel <= depth {
				allLinksMap[topLink] = httpParser.GetUniqueLinksFromUrl(topLink)
				for _, link := range allLinksMap[topLink] {
					if _, ok := visited[link]; !ok {
						q <- struct {
							string
							int
						}{link, topLevel + 1}
						visited[link] = true
						*size++
					}
				}

			}
			fmt.Println("#done go routine for link ", topLink, " and level ", topLevel)
		}(top.string, top.int, &size)

		if size == 0 {
			wg.Wait()
		}
	}

	return allLinksMap
}
