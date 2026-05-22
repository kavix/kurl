package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Options struct {
	Method  string
	URL     string
	Data    []byte
	Headers []string
	Timeout time.Duration
	Verbose bool
}

type RedirectHop struct {
	Status   string
	Method   string
	URL      string
	Location string
}

type Result struct {
	Request   *http.Request
	Response  *http.Response
	Redirects []RedirectHop
}

func Fetch(opts Options) (*Result, error) {
	if hasExplicitScheme(opts.URL) {
		return fetchSingle(opts, opts.URL)
	}

	return fetchConcurrentSchemes(opts)
}

func fetchConcurrentSchemes(opts Options) (*Result, error) {
	type outcome struct {
		result *Result
		err    error
		id     int
	}

	results := make(chan outcome, 2)
	cancels := make([]context.CancelFunc, 2)

	for id, candidate := range []string{"https://" + opts.URL, "http://" + opts.URL} {
		candidate := candidate
		id := id
		candCtx, candCancel := context.WithCancel(context.Background())
		cancels[id] = candCancel
		go func() {
			result, err := fetchSingleWithContext(candCtx, opts, candidate)
			results <- outcome{result: result, err: err, id: id}
		}()
	}

	var lastErr error
	for i := 0; i < 2; i++ {
		out := <-results
		if out.err == nil && out.result != nil {
			// Success! Cancel the other candidate (the loser)
			loserID := 1 - out.id
			cancels[loserID]()

			// Delay the winner's context cancellation until Close() is called on the response body.
			out.result.Response.Body = &cancelOnCloseReadCloser{
				ReadCloser: out.result.Response.Body,
				cancel:     cancels[out.id],
			}
			return out.result, nil
		}
		// If it failed, cancel its own context immediately
		cancels[out.id]()
		lastErr = out.err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unable to fetch %q", opts.URL)
	}
	return nil, lastErr
}

type cancelOnCloseReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnCloseReadCloser) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

func fetchSingle(opts Options, target string) (*Result, error) {
	ctx := context.Background()
	return fetchSingleWithContext(ctx, opts, target)
}

func fetchSingleWithContext(ctx context.Context, opts Options, target string) (*Result, error) {
	transport := tunedTransport()
	cli := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	headers, err := parseHeaders(opts.Headers)
	if err != nil {
		return nil, err
	}

	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 kurl/1.0")
	}
	if headers.Get("Accept") == "" {
		headers.Set("Accept", "*/*")
	}

	requestURL := target
	method := strings.ToUpper(opts.Method)
	if method == "" {
		method = http.MethodGet
	}
	body := opts.Data
	var redirects []RedirectHop

	for attempt := 0; attempt < 10; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		for name, values := range headers {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}

		resp, err := cli.Do(req)
		if err != nil {
			return nil, err
		}

		if location := resp.Header.Get("Location"); isRedirect(resp.StatusCode) && location != "" {
			redirects = append(redirects, RedirectHop{
				Status:   resp.Status,
				Method:   method,
				URL:      requestURL,
				Location: location,
			})
			nextURL, err := resolveURL(requestURL, location)
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
			nextMethod, nextBody := redirectRequest(method, resp.StatusCode, body)
			resp.Body.Close()
			requestURL = nextURL
			method = nextMethod
			body = nextBody
			continue
		}

		return &Result{Request: req, Response: resp, Redirects: redirects}, nil
	}

	return nil, fmt.Errorf("too many redirects")
}

func hasExplicitScheme(rawURL string) bool {
	return strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
}

func resolveHostConcurrent(ctx context.Context, host string) ([]net.IP, error) {
	if net.ParseIP(host) != nil {
		return []net.IP{net.ParseIP(host)}, nil
	}

	type outcome struct {
		ips []net.IP
		err error
	}

	results := make(chan outcome, 2)
	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// 1. Resolve using public Cloudflare DNS 1.1.1.1
	go func() {
		cfCtx, cfCancel := context.WithTimeout(ctxCancel, 800*time.Millisecond)
		defer cfCancel()

		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 500 * time.Millisecond}
				return d.DialContext(ctx, network, "1.1.1.1:53")
			},
		}
		ips, err := resolver.LookupIP(cfCtx, "ip", host)
		results <- outcome{ips: ips, err: err}
	}()

	// 2. Resolve using default System Resolver
	go func() {
		ips, err := net.DefaultResolver.LookupIP(ctxCancel, "ip", host)
		results <- outcome{ips: ips, err: err}
	}()

	var lastErr error
	for i := 0; i < 2; i++ {
		select {
		case out := <-results:
			if out.err == nil && len(out.ips) > 0 {
				cancel()
				return out.ips, nil
			}
			lastErr = out.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, lastErr
}

func tunedTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := resolveHostConcurrent(ctx, host)
			if err != nil {
				return dialer.DialContext(ctx, network, addr)
			}

			var dialErr error
			for _, ip := range ips {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if err == nil {
					return conn, nil
				}
				dialErr = err
			}
			return nil, dialErr
		},
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
}

func parseHeaders(values []string) (http.Header, error) {
	headers := http.Header{}
	for _, item := range values {
		name, value, ok := strings.Cut(item, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q", item)
		}
		headers.Add(strings.TrimSpace(name), strings.TrimSpace(value))
	}
	return headers, nil
}

func resolveURL(baseURL string, location string) (string, error) {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsedLocation, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return parsedBase.ResolveReference(parsedLocation).String(), nil
}

func redirectRequest(method string, statusCode int, body []byte) (string, []byte) {
	switch statusCode {
	case http.StatusSeeOther:
		return http.MethodGet, nil
	case http.StatusMovedPermanently, http.StatusFound:
		if method != http.MethodGet && method != http.MethodHead {
			return http.MethodGet, nil
		}
	}
	return method, body
}

func isRedirect(code int) bool {
	return code >= 300 && code < 400
}