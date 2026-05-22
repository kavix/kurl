package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"brew-terminal-curl/internal/response"
)

func main() {
	inputPath := flag.String("in", "", "read a raw HTTP response from a file instead of stdin")
	jsonOutput := flag.Bool("json", false, "render machine-readable JSON output")
	colorOutput := flag.Bool("color", true, "enable ANSI color in terminal output")
	method := flag.String("method", http.MethodGet, "HTTP method used when fetching a URL")
	flag.Parse()

	parsed, err := loadResponse(*inputPath, *method, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *jsonOutput {
		encoded, err := response.RenderJSON(parsed)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(encoded)
		return
	}

	fmt.Print(response.RenderText(parsed, *colorOutput))
}

func loadResponse(inputPath string, method string, args []string) (response.Response, error) {
	if inputPath != "" || len(args) == 0 {
		data, err := readInput(inputPath)
		if err != nil {
			return response.Response{}, err
		}

		return response.Parse(string(data)), nil
	}

	if len(args) > 1 {
		return response.Response{}, fmt.Errorf("expected a single URL argument")
	}

	parsedURL, err := url.ParseRequestURI(args[0])
	if err != nil {
		return response.Response{}, fmt.Errorf("invalid url %q: %w", args[0], err)
	}

	return fetchURL(parsedURL.String(), method)
}

func fetchURL(target string, method string) (response.Response, error) {
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return response.Response{}, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return response.Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Response{}, err
	}

	return response.FromHTTPResponse(resp.Status, resp.Header, body), nil
}

func readInput(path string) ([]byte, error) {
	if path == "" {
		return io.ReadAll(os.Stdin)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
