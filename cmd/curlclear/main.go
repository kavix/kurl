package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"brew-terminal-curl/internal/response"
)

func main() {
	inputPath := flag.String("in", "", "read a raw HTTP response from a file instead of stdin")
	jsonOutput := flag.Bool("json", false, "render machine-readable JSON output")
	colorOutput := flag.Bool("color", true, "enable ANSI color in terminal output")
	flag.Parse()

	data, err := readInput(*inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	parsed := response.Parse(string(data))
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