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

type SitesBench struct {
	wg      *sync.WaitGroup
	sites   []api.ResponseItem
	jobs    chan *RequestItem
	results chan *Result
}

func NewSitesBench(sites []api.ResponseItem) (*SitesBench, error) {
	conf := config.GetConfig()

	bench := &SitesBench{
		new(sync.WaitGroup),
		sites,
		make(chan *RequestItem, conf.RequestsPerHost*len(sites)),
		make(chan *Result, conf.RequestsPerHost*len(sites)),
	}

	tlsConf, err := cert.NewTLSConfig()
	if err != nil {
		return &SitesBench{}, err
	}

	client := NewHttpClient(tlsConf.Clone())

	for i := 0; i < conf.Concurrency; i++ {
		go NewHTTPWorker(bench.wg, client, tlsConf.Clone(), bench.jobs, bench.results).Run()
	}

	return bench, nil
}

func (s *SitesBench) Close() {
	close(s.jobs)
	close(s.results)
}

func (s *SitesBench) Run(ctx context.Context) (map[string]int, error) {
	benchResult := make(map[string]int)

	conf := config.GetConfig()

	var chunks [][]api.ResponseItem
	for i := 0; i < len(s.sites); i += conf.ChunkSize {
		end := i + conf.ChunkSize

		if end > len(s.sites) {
			end = len(s.sites)
		}

		chunks = append(chunks, s.sites[i:end])
	}

	fmt.Printf("Total chunks to process %d\n", len(chunks))

	// Precache connections
	if err := s.processChunk(ctx, s.sites, 1, benchResult); err != nil {
		return map[string]int{}, err
	}

	// Process sites in chunks
	for idx, chunk := range chunks {
		now := time.Now()
		if err := s.processChunk(ctx, chunk, conf.RequestsPerHost, benchResult); err != nil {
			break
		}
		fmt.Printf("chunk %d processed after %f seconds\n", idx+1, time.Since(now).Seconds())
	}

	return benchResult, nil
}

func (s *SitesBench) processChunk(
	ctx context.Context,
	sites []api.ResponseItem,
	requests int,
	benchResult map[string]int,
) error {
	totalSites := len(sites)
	for _, site := range sites {
		s.wg.Add(1)

		_, ok := benchResult[site.Host]
		if !ok {
			benchResult[site.Host] = 0
		}

		go func(wg *sync.WaitGroup, s api.ResponseItem, c chan<- *RequestItem) {
			for i := 0; i < requests; i++ {
				c <- &RequestItem{
					s,
				}
			}
			wg.Done()
		}(s.wg, site, s.jobs)
	}
	s.wg.Wait()

	total := totalSites * requests
	for i := 0; i < total; i++ {
		select {
		case res := <-s.results:
			if res.value {
				benchResult[res.host]++
			}
		case <-ctx.Done():
			return fmt.Errorf("context deadline exeeded")
		}
	}

	return nil
}
