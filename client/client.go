package client

import (
	"bytes"
	"context"
	"fmt"
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
		res, err := fetchSingle(opts, opts.URL, true)
		if err == nil {
			return res, nil
		}
		if isDNSFailure(err) {
			return fetchSingle(opts, opts.URL, false)
		}
		return nil, err
	}

	result, err := fetchConcurrentSchemes(opts, true)
	if err == nil {
		return result, nil
	}

	if isDNSFailure(err) {
		return fetchConcurrentSchemes(opts, false)
	}

	return nil, err
}

func fetchConcurrentSchemes(opts Options, usePublicDNS bool) (*Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type outcome struct {
		result *Result
		err    error
	}

	results := make(chan outcome, 2)
	for _, candidate := range []string{"https://" + opts.URL, "http://" + opts.URL} {
		candidate := candidate
		go func() {
			result, err := fetchSingleWithContext(ctx, opts, candidate, usePublicDNS)
			results <- outcome{result: result, err: err}
		}()
	}

	var lastErr error
	for i := 0; i < 2; i++ {
		outcome := <-results
		if outcome.err == nil && outcome.result != nil {
			cancel()
			return outcome.result, nil
		}
		lastErr = outcome.err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unable to fetch %q", opts.URL)
	}
	return nil, lastErr
}

func fetchSingle(opts Options, target string, usePublicDNS bool) (*Result, error) {
	ctx := context.Background()
	return fetchSingleWithContext(ctx, opts, target, usePublicDNS)
}

func fetchSingleWithContext(ctx context.Context, opts Options, target string, usePublicDNS bool) (*Result, error) {
	transport := tunedTransport(usePublicDNS)
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

func tunedTransport(usePublicDNS bool) *http.Transport {
	var resolver *net.Resolver
	if usePublicDNS {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 5 * time.Second}
				return d.DialContext(ctx, network, "1.1.1.1:53")
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  resolver,
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
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

func isDNSFailure(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "lookup") || strings.Contains(text, "dns") || strings.Contains(text, "read udp") || strings.Contains(text, "no such host")
}