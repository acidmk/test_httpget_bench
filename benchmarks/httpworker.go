package benchmarks

import (
	"crypto/tls"
	"net/http"
	"test-bench/config"
	"time"
)

type Result struct {
	host  string
	value bool
}

type HTTPWorker struct {
	client  *http.Client
	jobs    chan *http.Request
	results chan *Result
}

func NewHTTPWorker(conf *config.Config, jobs chan *http.Request, results chan *Result) *HTTPWorker {
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		DisableCompression: conf.UseCompression,
		DisableKeepAlives:  true,
		TLSClientConfig:    tlsconfig,
	}

	return &HTTPWorker{
		&http.Client{
			Transport: transport,
			Timeout:   time.Duration(conf.RequestTimeout) * time.Second,
		},
		jobs,
		results,
	}
}

func (h *HTTPWorker) Run() {
	for job := range h.jobs {
		result := <-h.doRequest(job)
		h.results <- result
	}
}

func (h *HTTPWorker) doRequest(request *http.Request) chan *Result {
	ch := make(chan *Result, 1)
	go func() {
		if r, err := h.client.Do(request); err != nil || r.StatusCode != http.StatusOK {
			ch <- &Result{request.Host, false}
		}
		ch <- &Result{request.Host, true}
	}()

	return ch
}
