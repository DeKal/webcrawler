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
	Mode           Mode
	MaxConcurrency int8
	EnableDebug    bool
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

func (w WebCrawler) getMaxConcurrency() int8 {
	if w.MaxConcurrency == 0 {
		return 16
	}
	return min(w.MaxConcurrency, 32)
}

func (w WebCrawler) getEnableDebug() bool {
	return w.EnableDebug
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

func (w WebCrawler) asyncCrawlWithChannel(rootLink string, depth int) map[string][]string {
	var (
		allLinksMap  = make(map[string][]string)
		linksChannel = make(chan struct {
			root  string
			links []string
		}, 100000)
		q = make(chan struct {
			string
			int
		}, 100000)
		size    int
		visited = make(map[string]bool)
		wg      sync.WaitGroup
		lock    sync.RWMutex
	)

	go func() {
		for item := range linksChannel {
			allLinksMap[item.root] = item.links
		}
	}()

	// Create a buffered channel to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, w.getMaxConcurrency())

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
			semaphore <- struct{}{}
			if w.getEnableDebug() {
				fmt.Println("exec go routine for link ", topLink, " and level ", topLevel)
				defer fmt.Println("#done go routine for link ", topLink, " and level ", topLevel)
			}

			defer wg.Done()
			defer func() { <-semaphore }()

			if topLevel <= depth {
				links := httpParser.GetUniqueLinksFromUrl(topLink)
				linksChannel <- struct {
					root  string
					links []string
				}{
					topLink, links,
				}
				for _, link := range links {

					lock.Lock()
					if _, ok := visited[link]; !ok {
						q <- struct {
							string
							int
						}{link, topLevel + 1}
						visited[link] = true
						*size++
					}
					lock.Unlock()
				}

			}
		}(top.string, top.int, &size)

		if size == 0 {
			wg.Wait()
		}
	}

	close(linksChannel)

	return allLinksMap
}
