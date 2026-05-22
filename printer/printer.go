package printer

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"brew-terminal-curl/client"
	"brew-terminal-curl/color"
)

type Options struct {
	Color       bool
	Raw         bool
	HeadersOnly bool
	BodyOnly    bool
	Verbose     bool
	OutputPath  string
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
	return renderBody(w, result, opts.Color, opts.OutputPath)
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
	return renderBody(w, result, opts.Color, opts.OutputPath)
}

func renderBody(w io.Writer, result *client.Result, enabled bool, outputPath string) error {
	contentType := result.Response.Header.Get("Content-Type")
	writer := w
	if outputPath != "" {
		file, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = io.MultiWriter(w, file)
	}

	if isBinary(contentType) {
		_, err := fmt.Fprintln(writer, "[Binary data - use -o to save]")
		return err
	}

	if isJSON(contentType, result.Response.ContentLength) {
		written, err := PrettyJSON(writer, result.Response.Body, enabled)
		if err != nil {
			return err
		}
		if written == 0 {
			_, err = fmt.Fprintln(writer, "No body")
		}
		return err
	}

	if isHTML(contentType) {
		written, err := PrettyHTML(writer, result.Response.Body, enabled)
		if err != nil {
			return err
		}
		if written == 0 {
			_, err = fmt.Fprintln(writer, "No body")
		}
		return err
	}

	written, err := io.Copy(writer, result.Response.Body)
	if err != nil {
		return err
	}
	if written == 0 {
		_, err = fmt.Fprintln(writer, "No body")
		return err
	}
	_, err = fmt.Fprintln(writer)
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
