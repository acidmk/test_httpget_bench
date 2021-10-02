package benchmarks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"test-bench/api"
	"test-bench/cert"
	"test-bench/config"
	"time"
)

func MeasureMaxRequestsForSites(ctx context.Context, sites []api.ResponseItem) (map[string]int, error) {
	benchResult := make(map[string]int)

	conf := config.GetConfig()

	jobs := make(chan *RequestItem, conf.RequestsPerHost*len(sites))
	results := make(chan *Result, conf.RequestsPerHost*len(sites))
	stop := make(chan struct{})

	tlsConf, err := cert.NewTLSConfig()
	if err != nil {
		return map[string]int{}, err
	}
	for i := 0; i < conf.Concurrency; i++ {
		go NewHTTPWorker(stop, tlsConf.Clone(), jobs, results).Run()
	}

	var wg sync.WaitGroup
	var rWg sync.WaitGroup

	var chunks [][]api.ResponseItem
	chunkSize := len(sites) / 4
	for i := 0; i < len(sites); i += chunkSize {
		end := i + chunkSize

		if end > len(sites) {
			end = len(sites)
		}

		chunks = append(chunks, sites[i:end])
	}

	fmt.Printf("Total chunks to process %d\n", len(chunks))

	for idx, chunk := range chunks {
		totalSites := len(chunk)
		now := time.Now()
		for _, site := range chunk {
			wg.Add(1)

			benchResult[site.Host] = 0
			go func(wg *sync.WaitGroup, rWg *sync.WaitGroup, s api.ResponseItem, c chan<- *RequestItem) {
				_, err := net.LookupIP(s.Host)
				if err != nil {
					fmt.Println(err)
					totalSites--
					wg.Done()
					return
				}

				for i := 0; i < conf.RequestsPerHost; i++ {
					c <- &RequestItem{
						s,
					}
				}
				wg.Done()
			}(&wg, &rWg, site, jobs)
		}

		wg.Wait()

		time.Since(now).Seconds()

		total := totalSites * conf.RequestsPerHost
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

		fmt.Printf("chunk %d processed after %f seconds\n", idx+1, time.Since(now).Seconds())
	}

	close(jobs)
	stop <- struct{}{}

	return benchResult, nil
}
