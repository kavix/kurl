package printer

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"brew-terminal-curl/color"
	"golang.org/x/net/html"
)

// PrettyHTML parses HTML from r and writes a pretty-printed, indented, colorized version to w.
func PrettyHTML(w io.Writer, r io.Reader, enabled bool) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}

	contentStr := string(data)
	hasHtml := strings.Contains(strings.ToLower(contentStr), "<html")
	hasBody := strings.Contains(strings.ToLower(contentStr), "<body")
	hasHead := strings.Contains(strings.ToLower(contentStr), "<head")

	cw := &countingWriter{w: w}
	if err := format(cw, doc, 0, enabled, hasHtml, hasBody, hasHead); err != nil {
		return cw.count, err
	}

	return cw.count, nil
}

type countingWriter struct {
	w     io.Writer
	count int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.count += int64(n)
	return n, err
}

func format(cw *countingWriter, n *html.Node, depth int, enabled bool, hasHtml, hasBody, hasHead bool) error {
	switch n.Type {
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := format(cw, c, depth, enabled, hasHtml, hasBody, hasHead); err != nil {
				return err
			}
		}
	case html.DoctypeNode:
		doctype := renderDoctype(n, enabled)
		if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+doctype+"\n"); err != nil {
			return err
		}
	case html.CommentNode:
		comment := renderComment(n, enabled)
		if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+comment+"\n"); err != nil {
			return err
		}
	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+html.EscapeString(text)+"\n"); err != nil {
				return err
			}
		}
	case html.ElementNode:
		// Smart skip for auto-generated <html>, <head>, <body> tags
		if n.Data == "html" && !hasHtml {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := format(cw, c, depth, enabled, hasHtml, hasBody, hasHead); err != nil {
					return err
				}
			}
			return nil
		}
		if n.Data == "head" && !hasHead {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := format(cw, c, depth, enabled, hasHtml, hasBody, hasHead); err != nil {
					return err
				}
			}
			return nil
		}
		if n.Data == "body" && !hasBody {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := format(cw, c, depth, enabled, hasHtml, hasBody, hasHead); err != nil {
					return err
				}
			}
			return nil
		}

		// Format script and style block content as raw but indented text block
		if n.Data == "script" || n.Data == "style" {
			start := renderStartTag(n, enabled)
			if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+start+"\n"); err != nil {
				return err
			}
			if n.FirstChild != nil {
				lines := strings.Split(n.FirstChild.Data, "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" {
						if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth+1)+trimmed+"\n"); err != nil {
							return err
						}
					}
				}
			}
			end := renderEndTag(n, enabled)
			if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+end+"\n"); err != nil {
				return err
			}
			return nil
		}

		// Handle empty elements
		if n.FirstChild == nil {
			var tagStr string
			if isVoidElement(n.Data) {
				tagStr = renderVoidTag(n, enabled)
			} else {
				tagStr = renderStartTag(n, enabled) + renderEndTag(n, enabled)
			}
			if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+tagStr+"\n"); err != nil {
				return err
			}
			return nil
		}

		// Format simple inline elements (and their inline children) on a single line
		if hasOnlyInlineChildren(n) {
			var sb strings.Builder
			sb.WriteString(renderStartTag(n, enabled))
			if err := formatInline(&sb, n, enabled); err != nil {
				return err
			}
			sb.WriteString(renderEndTag(n, enabled))

			if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+sb.String()+"\n"); err != nil {
				return err
			}
			return nil
		}

		// Otherwise, a standard block-level element
		start := renderStartTag(n, enabled)
		if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+start+"\n"); err != nil {
			return err
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode && strings.TrimSpace(c.Data) == "" {
				continue
			}
			if err := format(cw, c, depth+1, enabled, hasHtml, hasBody, hasHead); err != nil {
				return err
			}
		}

		end := renderEndTag(n, enabled)
		if _, err := fmt.Fprint(cw, strings.Repeat("  ", depth)+end+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func formatInline(sb *strings.Builder, n *html.Node, enabled bool) error {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			sb.WriteString(html.EscapeString(c.Data))
		case html.CommentNode:
			sb.WriteString(renderComment(c, enabled))
		case html.ElementNode:
			if c.FirstChild == nil {
				if isVoidElement(c.Data) {
					sb.WriteString(renderVoidTag(c, enabled))
				} else {
					sb.WriteString(renderStartTag(c, enabled) + renderEndTag(c, enabled))
				}
			} else {
				sb.WriteString(renderStartTag(c, enabled))
				if err := formatInline(sb, c, enabled); err != nil {
					return err
				}
				sb.WriteString(renderEndTag(c, enabled))
			}
		}
	}
	return nil
}

func isVoidElement(tagName string) bool {
	switch strings.ToLower(tagName) {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

func isInlineElement(tagName string) bool {
	switch strings.ToLower(tagName) {
	case "a", "abbr", "acronym", "b", "bdo", "big", "br", "button", "cite", "code", "dfn", "em", "i", "img", "input", "kbd", "label", "map", "object", "output", "q", "samp", "script", "select", "small", "span", "strong", "sub", "sup", "textarea", "time", "tt", "var":
		return true
	}
	return false
}

func hasOnlyInlineChildren(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if !isInlineElement(c.Data) || !hasOnlyInlineChildren(c) {
				return false
			}
		}
	}
	return true
}

func renderStartTag(n *html.Node, enabled bool) string {
	var sb strings.Builder
	sb.WriteString(color.Border(enabled, "<"))
	sb.WriteString(color.Key(enabled, n.Data))
	for _, attr := range n.Attr {
		sb.WriteString(" ")
		sb.WriteString(color.Number(enabled, attr.Key))
		sb.WriteString(color.Border(enabled, "="))
		sb.WriteString(color.String(enabled, `"`+html.EscapeString(attr.Val)+`"`))
	}
	sb.WriteString(color.Border(enabled, ">"))
	return sb.String()
}

func renderEndTag(n *html.Node, enabled bool) string {
	var sb strings.Builder
	sb.WriteString(color.Border(enabled, "</"))
	sb.WriteString(color.Key(enabled, n.Data))
	sb.WriteString(color.Border(enabled, ">"))
	return sb.String()
}

func renderVoidTag(n *html.Node, enabled bool) string {
	var sb strings.Builder
	sb.WriteString(color.Border(enabled, "<"))
	sb.WriteString(color.Key(enabled, n.Data))
	for _, attr := range n.Attr {
		sb.WriteString(" ")
		sb.WriteString(color.Number(enabled, attr.Key))
		sb.WriteString(color.Border(enabled, "="))
		sb.WriteString(color.String(enabled, `"`+html.EscapeString(attr.Val)+`"`))
	}
	sb.WriteString(color.Border(enabled, " />"))
	return sb.String()
}

func renderComment(n *html.Node, enabled bool) string {
	return color.Border(enabled, "<!--"+n.Data+"-->")
}

func renderDoctype(n *html.Node, enabled bool) string {
	return color.Wrap(enabled, color.Bold+color.Magenta, "<!DOCTYPE "+n.Data+">")
}
