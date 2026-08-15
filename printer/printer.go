package printer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kavix/kurl/client"
	"github.com/kavix/kurl/color"
	"github.com/kavix/kurl/internal/filter"
)

type Options struct {
	Color         bool
	Raw           bool
	HeadersOnly   bool
	BodyOnly      bool
	Verbose       bool
	OutputPath    string
	FilterQuery   string
	FilterKeys    string
	FilterFlatten bool
}

func Render(w io.Writer, result *client.Result, opts Options, elapsed time.Duration) error {
	if opts.Raw {
		return renderRaw(w, result, opts)
	}

	statusLine := fmt.Sprintf("kurl · %s %s", result.Request.Method, result.Request.URL.String())
	if _, err := fmt.Fprintln(w, boxTop(opts.Color, statusLine)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, boxBottom(opts.Color, len(statusLine)+4)); err != nil {
		return err
	}

	if opts.BodyOnly {
		return renderBodyOnly(w, result, opts)
	}

	if opts.HeadersOnly {
		return renderHeadersOnly(w, result, opts)
	}

	if err := renderSummary(w, result, opts.Color, elapsed); err != nil {
		return err
	}

	if opts.Verbose {
		if err := renderVerbose(w, result, opts.Color); err != nil {
			return err
		}
	}

	if err := renderHeadersAndBody(w, result, opts); err != nil {
		return err
	}

	return nil
}

func renderRaw(w io.Writer, result *client.Result, opts Options) error {
	if opts.OutputPath != "" {
		return saveBodyToFile(w, result.Response.Body, opts.OutputPath)
	}
	_, err := io.Copy(w, result.Response.Body)
	return err
}

func renderSummary(w io.Writer, result *client.Result, enabled bool, elapsed time.Duration) error {
	status := result.Response.StatusCode
	proto := result.Response.Proto
	if proto == "" {
		proto = "HTTP/1.1"
	}
	if _, err := fmt.Fprintf(w, "  %-8s %s\n", color.Title(enabled, "STATUS"), color.Status(enabled, status, result.Response.Status)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-8s %s\n", color.Title(enabled, "TIME"), elapsed.Round(time.Millisecond)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-8s %s\n", color.Title(enabled, "PROTO"), proto); err != nil {
		return err
	}
	return nil
}

func renderVerbose(w io.Writer, result *client.Result, enabled bool) error {
	if len(result.Redirects) > 0 {
		if _, err := fmt.Fprintln(w, sectionTitle(enabled, "REDIRECTS")); err != nil {
			return err
		}
		for _, hop := range result.Redirects {
			if _, err := fmt.Fprintf(w, "  %s %s -> %s\n", color.Header(enabled, hop.Status), hop.Method, hop.Location); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w, sectionTitle(enabled, "REQUEST")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s %s\n", color.Header(enabled, result.Request.Method), result.Request.URL.String()); err != nil {
		return err
	}
	for name, values := range result.Request.Header {
		for _, value := range values {
			if _, err := fmt.Fprintf(w, "  %s %s\n", color.Header(enabled, name), value); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderHeadersAndBody(w io.Writer, result *client.Result, opts Options) error {
	if _, err := fmt.Fprintln(w, sectionTitle(opts.Color, "HEADERS")); err != nil {
		return err
	}
	for name, values := range result.Response.Header {
		for _, value := range values {
			if _, err := fmt.Fprintf(w, "  %-18s %s\n", color.Header(opts.Color, name), value); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w, sectionTitle(opts.Color, "BODY")); err != nil {
		return err
	}
	return renderBody(w, result, opts)
}

func renderHeadersOnly(w io.Writer, result *client.Result, opts Options) error {
	if _, err := fmt.Fprintln(w, sectionTitle(opts.Color, "HEADERS")); err != nil {
		return err
	}
	for name, values := range result.Response.Header {
		for _, value := range values {
			if _, err := fmt.Fprintf(w, "  %-18s %s\n", color.Header(opts.Color, name), value); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderBodyOnly(w io.Writer, result *client.Result, opts Options) error {
	if _, err := fmt.Fprintln(w, sectionTitle(opts.Color, "BODY")); err != nil {
		return err
	}
	return renderBody(w, result, opts)
}

func renderBody(w io.Writer, result *client.Result, opts Options) error {
	contentType := result.Response.Header.Get("Content-Type")
	enabled := opts.Color
	outputPath := opts.OutputPath

	if outputPath != "" && (opts.FilterQuery == "" && opts.FilterKeys == "" && !opts.FilterFlatten) {
		return saveBodyToFile(w, result.Response.Body, outputPath)
	}

	if isBinary(contentType) {
		_, err := fmt.Fprintln(w, "[Binary data - use -o to save]")
		return err
	}

	if isJSON(contentType, result.Response.ContentLength) {
		var bodyReader io.Reader = result.Response.Body
		if opts.FilterQuery != "" || opts.FilterKeys != "" || opts.FilterFlatten {
			data, err := io.ReadAll(result.Response.Body)
			if err != nil {
				return err
			}
			if opts.FilterQuery != "" {
				data, err = filter.ApplyFilter(data, opts.FilterQuery)
				if err != nil {
					return fmt.Errorf("filter error: %w", err)
				}
			}
			if opts.FilterKeys != "" {
				data, err = filter.FilterKeys(data, opts.FilterKeys)
				if err != nil {
					return fmt.Errorf("filter keys error: %w", err)
				}
			}
			if opts.FilterFlatten {
				data, err = filter.FlattenArray(data)
				if err != nil {
					return fmt.Errorf("flatten error: %w", err)
				}
			}
			bodyReader = bytes.NewReader(data)
		}

		if outputPath != "" {
			return saveBodyToFile(w, bodyReader, outputPath)
		}

		written, err := PrettyJSON(w, bodyReader, enabled)
		if err != nil {
			return err
		}
		if written == 0 {
			_, err = fmt.Fprintln(w, "No body")
		}
		return err
	}

	if isHTML(contentType) {
		written, err := PrettyHTML(w, result.Response.Body, enabled)
		if err != nil {
			return err
		}
		if written == 0 {
			_, err = fmt.Fprintln(w, "No body")
		}
		return err
	}

	written, err := io.Copy(w, result.Response.Body)
	if err != nil {
		return err
	}
	if written == 0 {
		_, err = fmt.Fprintln(w, "No body")
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}

func saveBodyToFile(w io.Writer, body io.Reader, outputPath string) error {
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "Saved response body to %s\n", outputPath)
	return err
}

func isJSON(contentType string, length int64) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json")
}

func isBinary(contentType string) bool {
	contentType = strings.ToLower(contentType)
	if contentType == "" {
		return false
	}
	return !strings.HasPrefix(contentType, "text/") && !strings.Contains(contentType, "/json")
}

func isHTML(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml+xml")
}

// RenderTiming prints the per-phase request timing breakdown. The client
// measures everything up to the first response byte; contentTransfer and total
// are supplied by the caller, which owns the wall-clock window around the body
// read.
func RenderTiming(w io.Writer, t *client.Timing, contentTransfer, total time.Duration, enabled bool) error {
	if _, err := fmt.Fprintln(w, sectionTitle(enabled, "TIMING")); err != nil {
		return err
	}
	rows := []struct {
		label string
		value time.Duration
	}{
		{"DNS Lookup", t.DNSLookup},
		{"TCP Connection", t.TCPConnection},
		{"TLS Handshake", t.TLSHandshake},
		{"Server Processing", t.ServerProcessing},
		{"Content Transfer", contentTransfer},
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "  %-18s %s\n", color.Header(enabled, row.label), row.value.Round(time.Millisecond)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "  %-18s %s\n", color.Title(enabled, "Total"), total.Round(time.Millisecond)); err != nil {
		return err
	}
	return nil
}

func sectionTitle(enabled bool, title string) string {
	return color.Title(enabled, "── "+title+" ──────────────────────────────────────")
}

func boxTop(enabled bool, title string) string {
	width := len(title) + 4
	return color.Border(enabled, "┌"+strings.Repeat("─", width)+"┐") + "\n" +
		color.Border(enabled, "│  "+title+"  │")
}

func boxBottom(enabled bool, width int) string {
	return color.Border(enabled, "└"+strings.Repeat("─", width)+"┘")
}
