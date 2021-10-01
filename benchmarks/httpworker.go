package benchmarks

import (
	"net"
	"time"
)

type RequestItem struct {
	Timeout time.Duration
	Host    string
	Ip      net.IP
}

type Result struct {
	host  string
	value bool
}

type HTTPWorker struct {
	jobs    chan *RequestItem
	results chan *Result
}

func NewHTTPWorker(jobs chan *RequestItem, results chan *Result) *HTTPWorker {
	return &HTTPWorker{
		jobs,
		results,
	}
}

func (h *HTTPWorker) Run() {
	for job := range h.jobs {
		select {
		case result := <-h.doRequest(job):
			h.results <- result
		}
	}
}

func (h *HTTPWorker) doRequest(request *RequestItem) chan *Result {
	ch := make(chan *Result, 1)
	go func() {
		conn, err := net.DialTimeout("tcp", request.Ip.String()+":https", request.Timeout)
		if err != nil {
			h.results <- &Result{request.Host, false}
			return
		}
		h.results <- &Result{request.Host, true}

		conn.Close()
	}()

	return ch
}
