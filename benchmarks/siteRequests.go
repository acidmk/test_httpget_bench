package benchmarks

import (
	"context"
	"fmt"
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

	tlsConf, err := cert.NewTLSConfig()
	if err != nil {
		return map[string]int{}, err
	}
	var wg sync.WaitGroup

	client := NewHttpClient(tlsConf.Clone())

	for i := 0; i < conf.Concurrency; i++ {
		go NewHTTPWorker(&wg, client, tlsConf.Clone(), jobs, results).Run()
	}

	var chunks [][]api.ResponseItem
	chunkSize := conf.ChunkSize
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
			go func(wg *sync.WaitGroup, s api.ResponseItem, c chan<- *RequestItem) {
				for i := 0; i < conf.RequestsPerHost; i++ {
					c <- &RequestItem{
						s,
					}
				}
				wg.Done()
			}(&wg, site, jobs)
		}

		wg.Wait()

		total := totalSites * conf.RequestsPerHost
		for i := 0; i < total; i++ {
			select {
			case res := <-results:
				if res.value {
					benchResult[res.host]++
				}
			case <-ctx.Done():
				return benchResult, nil
			}
		}

		fmt.Printf("chunk %d processed after %f seconds\n", idx+1, time.Since(now).Seconds())
	}

	close(jobs)

	return benchResult, nil
}
