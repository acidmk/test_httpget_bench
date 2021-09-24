package benchmarks

import (
	"context"
	"errors"
	"net/http"
	"test-bench/api"
	"test-bench/config"
)

func MeasureMaxRequestsForSites(ctx context.Context, sites []api.ResponseItem) (map[string]int, error) {
	benchResult := make(map[string]int)

	conf := config.GetConfig()

	jobs := make(chan *http.Request, conf.Concurrency)
	results := make(chan *Result, conf.RequestsPerHost*len(sites))

	for i := 0; i < conf.Concurrency; i++ {
		go NewHTTPWorker(conf, jobs, results).Run()
	}

	for _, site := range sites {
		benchResult[site.Host] = 0

		req, _ := http.NewRequest(http.MethodGet, site.Url, nil)
		for i := 0; i < conf.RequestsPerHost; i++ {
			reqCpy := req.Clone(context.Background())
			jobs <- reqCpy
		}
	}

	close(jobs)

	total := conf.RequestsPerHost * len(sites)
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
