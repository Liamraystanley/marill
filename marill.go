package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	urlsToCheck := os.Args[1:]
	results := []*Results{}
	var wg sync.WaitGroup

	// loop through all supplied urls and send them to a worker to be fetched
	for _, url := range urlsToCheck {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			fmt.Printf("[\033[1;36m---\033[0;m] [\033[0;32m------\033[0;m] \033[0;95mStarting to scan %s\033[0;m\n", url)
			result := Crawl(url, "")
			results = append(results, result)

			fmt.Printf("[\033[1;36m---\033[0;m] [\033[0;32m%4dms\033[0;m] \033[0;95mFinished scanning %s\033[0;m\n", result.TotalTime.Milli, url)
		}(url)
	}

	// wait for all workers to complete their tasks
	wg.Wait()

	// print out cool stuff here based on results
}
