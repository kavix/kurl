package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"brew-terminal-curl/client"
	"brew-terminal-curl/printer"
)

func main() {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Testing kurl HTML Formatter</title>
		<meta charset="utf-8" />
	</head>
	<body>
		<!-- Introduction Section -->
		<div class="content" id="main-content">
			<h1>Welcome to <b>kurl</b> formatting!</h1>
			<p>This is a paragraph with <i>italics</i>, <b>bold text</b>, and a <a href="https://github.com">link</a>.</p>
			<br />
			<ul class="features-list">
				<li>Auto indentation</li>
				<li>Tag colorization</li>
				<li>Smart inline tag collapsing</li>
			</ul>
		</div>
		<script>
			console.log("Formatting is works!");
			if (true) {
				alert("Success!");
			}
		</script>
	</body>
	</html>
	`

	reqURL, _ := url.Parse("https://localhost/test")
	res := &client.Result{
		Request: &http.Request{
			Method: "GET",
			URL:    reqURL,
		},
		Response: &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Proto:      "HTTP/2.0",
			Header: http.Header{
				"Content-Type": []string{"text/html; charset=utf-8"},
			},
			Body: io.NopCloser(bytes.NewReader([]byte(htmlContent))),
		},
	}

	opts := printer.Options{
		Color:       true, // Force enable color for preview
		Raw:         false,
		HeadersOnly: false,
		BodyOnly:    false,
		Verbose:     false,
	}

	fmt.Println("=== RUNNING VISUAL VERIFICATION ===")
	err := printer.Render(os.Stdout, res, opts, 45*time.Millisecond)
	if err != nil {
		fmt.Printf("Error rendering: %v\n", err)
	}
	fmt.Println("=== END VISUAL VERIFICATION ===")
}
