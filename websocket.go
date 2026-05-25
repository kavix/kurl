package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"brew-terminal-curl/color"
	"brew-terminal-curl/printer"
	"golang.org/x/net/websocket"
)

func runWebSocket(opts cliOptions) {
	useColor := color.AutoEnabled(os.Stdout) && !opts.noColor

	// Ensure the URL has ws:// or wss:// scheme
	if !strings.HasPrefix(opts.url, "ws://") && !strings.HasPrefix(opts.url, "wss://") {
		fatal(fmt.Errorf("invalid websocket url %q (must start with ws:// or wss://)", opts.url))
	}

	config, err := websocket.NewConfig(opts.url, opts.url)
	if err != nil {
		fatal(fmt.Errorf("failed to create websocket configuration: %w", err))
	}

	// Attach headers from options if any
	if len(opts.headers) > 0 {
		headers := http.Header{}
		for _, item := range opts.headers {
			name, value, ok := strings.Cut(item, ":")
			if !ok {
				fatal(fmt.Errorf("invalid header %q", item))
			}
			headers.Add(strings.TrimSpace(name), strings.TrimSpace(value))
		}
		config.Header = headers
	}

	// Inject standard User-Agent and Accept headers if not custom set
	if config.Header.Get("User-Agent") == "" {
		config.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 kurl/1.0")
	}
	if config.Header.Get("Accept") == "" {
		config.Header.Set("Accept", "*/*")
	}

	ws, err := websocket.DialConfig(config)
	if err != nil {
		fatal(fmt.Errorf("failed to connect to websocket: %w", err))
	}
	defer ws.Close()

	// Print connection header
	statusLine := fmt.Sprintf("kurl · WebSocket Connected to %s", opts.url)
	fmt.Fprintln(os.Stdout, boxTop(useColor, statusLine))
	fmt.Fprintln(os.Stdout, boxBottom(useColor, len(statusLine)+4))
	fmt.Fprintln(os.Stdout, color.Wrap(useColor, color.Dim, "Type messages and press Enter to send. Press Ctrl+C to exit.\n"))

	// Receive goroutine
	go func() {
		var msg string
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				if err == io.EOF {
					fmt.Fprintln(os.Stdout, color.Wrap(useColor, color.Bold+color.Red, "\nDisconnected by remote host."))
				} else if !strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Fprintf(os.Stderr, "\nError receiving: %v\n", err)
				}
				os.Exit(0)
			}

			printedMsg := msg
			if isJSONString(msg) {
				var buf bytes.Buffer
				_, err := printer.PrettyJSON(&buf, strings.NewReader(msg), useColor)
				if err == nil {
					printedMsg = "\n" + strings.TrimSpace(buf.String())
				}
			}

			prefix := color.Wrap(useColor, color.Bold+color.Green, "[RECV] <")
			fmt.Fprintf(os.Stdout, "%s %s\n", prefix, printedMsg)
		}
	}()

	// Send loop
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		err := websocket.Message.Send(ws, text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
			break
		}

		prefix := color.Wrap(useColor, color.Bold+color.Cyan, "[SEND] >")
		fmt.Fprintf(os.Stdout, "%s %s\n", prefix, text)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func isJSONString(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

func boxTop(enabled bool, title string) string {
	width := len(title) + 4
	return color.Border(enabled, "┌"+strings.Repeat("─", width)+"┐") + "\n" +
		color.Border(enabled, "│  "+title+"  │")
}

func boxBottom(enabled bool, width int) string {
	return color.Border(enabled, "└"+strings.Repeat("─", width)+"┘")
}
