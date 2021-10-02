package benchmarks

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"test-bench/api"
	"test-bench/config"
	"time"
)

type RequestItem struct {
	Site api.ResponseItem
}

type Result struct {
	host  string
	value bool
}

type HTTPWorker struct {
	wg      *sync.WaitGroup
	dialer  *tls.Dialer
	client  *http.Client
	jobs    chan *RequestItem
	results chan *Result
}

func NewHTTPWorker(
	wg *sync.WaitGroup,
	httpClient *http.Client,
	tlsConf *tls.Config,
	jobs chan *RequestItem,
	results chan *Result,
) *HTTPWorker {
	return &HTTPWorker{
		wg,
		&tls.Dialer{
			&net.Dialer{},
			tlsConf,
		},
		httpClient,
		jobs,
		results,
	}
}

func NewHttpClient(tlsConf *tls.Config) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 10000
	t.MaxConnsPerHost = 10000
	t.MaxIdleConnsPerHost = 10000
	t.DisableCompression = false
	t.TLSClientConfig = tlsConf

	return &http.Client{
		Timeout:   time.Duration(config.GetConfig().RequestTimeout) * time.Second,
		Transport: t,
	}
}

func (h *HTTPWorker) Run() {
	for job := range h.jobs {
		h.wg.Wait()
		result := <-h.doRequest(job)
		h.results <- result
	}
}

func (h *HTTPWorker) doRequest(request *RequestItem) chan *Result {
	ch := make(chan *Result, 1)
	go func() {
		conf := config.GetConfig()

		if conf.UseHttpGet {
			req, _ := http.NewRequest("GET", request.Site.Url, nil)
			res, err := h.client.Do(req)

			if err != nil || res.StatusCode != http.StatusOK {
				ch <- &Result{request.Site.Host, false}
				return
			}
			defer res.Body.Close()

			ch <- &Result{request.Site.Host, true}
			return
		}

		// in case www is cut from the host
		u, _ := url.Parse(request.Site.Url)
		deadline := time.Now().Add(time.Second * time.Duration(conf.RequestTimeout))
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		port := 443
		if strings.Contains(request.Site.Url, "http://") {
			fmt.Println("http site "+request.Site.Host)
			port = 80
		}
		conn, err := h.dialer.DialContext(ctx, "tcp", net.JoinHostPort(u.Host, strconv.Itoa(port)))
		if err != nil {
			h.results <- &Result{request.Site.Host, false}
			return
		}
		defer conn.Close()

		msg := fmt.Sprintf("GET %s HTTP/1.0\r\n"+
			"Host: %s\r\n"+
			"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\n"+
			"User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0\r\n"+
			"Accept-Encoding: gzip, deflate, br\r\n"+
			"Connection: close\r\n"+
			"\r\n\r\n",
			u.RequestURI(),
			u.Hostname(),
		)
		conn.SetDeadline(deadline)
		conn.Write([]byte(msg))
		status, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil || !strings.Contains(status, "200") {
			ch <- &Result{request.Site.Host, false}
			return
		}

		ch <- &Result{request.Site.Host, true}
	}()

	return ch
}
