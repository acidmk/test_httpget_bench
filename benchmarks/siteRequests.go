package benchmarks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"test-bench/api"
	"test-bench/config"
	"time"
)

type RequestItem struct {
	Timeout time.Duration
	Host string
	Ip net.IP
}

func MeasureMaxRequestsForSites(ctx context.Context, sites []api.ResponseItem) (map[string]int, error) {
	benchResult := make(map[string]int)

	conf := config.GetConfig()

	jobs := make(chan *RequestItem, conf.Concurrency)
	results := make(chan *Result, conf.RequestsPerHost*len(sites))

	for i := 0; i < conf.Concurrency; i++ {
		go NewHTTPWorker(jobs, results).Run()
	}

	var wg sync.WaitGroup
	totalSites := len(sites)

	for _, site := range sites {
		wg.Add(1)
		benchResult[site.Host] = 0
		go func(wg *sync.WaitGroup, s api.ResponseItem, c chan<- *RequestItem) {
			ips, err := net.LookupIP(s.Host)
			if err != nil {
				fmt.Println(err)
				totalSites--
				wg.Done()
				return
			}
			timeout := time.Duration(conf.RequestTimeout) * time.Second

			for i := 0; i < conf.RequestsPerHost; i++ {
				c <- &RequestItem{
					timeout,
					s.Host,
					ips[i % len(ips)],
				}
			}
			wg.Done()
		}(&wg, site, jobs)
	}

	wg.Wait()

	close(jobs)

	total := totalSites*conf.RequestsPerHost
	for i := 0; i < total; i++ {
		select {
		case res := <-results:
			if res.value {
				benchResult[res.host]++
			}
		case <-ctx.Done():
			return map[string]int{}, errors.New("context deadline exceeded")
		}
	}

	return benchResult, nil
}
