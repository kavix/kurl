package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Response struct {
	Mode           string   `json:"mode"`
	StatusLine     string   `json:"status_line,omitempty"`
	Headers        []Header `json:"headers,omitempty"`
	Body           string   `json:"body,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	RequestLine    string   `json:"request_line,omitempty"`
	RequestHeaders []Header `json:"request_headers,omitempty"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func Parse(input string) Response {
	cleaned := strings.ReplaceAll(input, "\r\n", "\n")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return Response{Mode: "empty", Warnings: []string{"input is empty"}}
	}

	lines := strings.Split(cleaned, "\n")
	if looksVerbose(lines) {
		return parseVerbose(lines)
	}

	response := Response{Mode: "raw"}
	index := 0

	if strings.HasPrefix(lines[0], "HTTP/") {
		response.StatusLine = strings.TrimSpace(lines[0])
		index = 1
	} else {
		response.Warnings = append(response.Warnings, "missing HTTP status line")
	}

	for index < len(lines) {
		line := strings.TrimSpace(lines[index])
		if line == "" {
			index++
			break
		}

		name, value, found := strings.Cut(line, ":")
		if !found {
			break
		}

		response.Headers = append(response.Headers, Header{
			Name:  strings.TrimSpace(name),
			Value: strings.TrimSpace(value),
		})
		index++
	}

	if index < len(lines) {
		response.Body = strings.Join(lines[index:], "\n")
	}

	if response.StatusLine == "" && len(response.Headers) == 0 && response.Body == "" {
		response.Warnings = append(response.Warnings, "unable to detect response sections")
	}

	return response
}

func FromHTTPResponse(statusLine string, headers http.Header, body []byte) Response {
	response := Response{Mode: "http", StatusLine: statusLine}

	for name, values := range headers {
		for _, value := range values {
			response.Headers = append(response.Headers, Header{Name: name, Value: value})
		}
	}

	response.Body = FormatBody(body, headers.Get("Content-Type"))
	if response.Body == "" {
		response.Warnings = append(response.Warnings, "response body is empty")
	}

	return response
}

func FormatBody(body []byte, contentType string) string {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return ""
	}

	if isJSONContentType(contentType) || json.Valid(trimmed) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, trimmed, "", "  "); err == nil {
			return pretty.String()
		}
	}

	return string(body)
}

func RenderJSON(response Response) (string, error) {
	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes) + "\n", nil
}

func RenderText(response Response, color bool) string {
	var builder strings.Builder
	writeSectionHeader(&builder, "Status", color, colorCyan)
	if response.StatusLine != "" {
		builder.WriteString(style(response.StatusLine, color, colorGreen))
		builder.WriteString("\n")
	} else {
		builder.WriteString("(not detected)\n")
	}

	writeSectionHeader(&builder, "Headers", color, colorCyan)
	if len(response.Headers) == 0 {
		builder.WriteString("(none)\n")
	} else {
		for _, header := range response.Headers {
			builder.WriteString(fmt.Sprintf("%s: %s\n", style(header.Name, color, colorYellow), header.Value))
		}
	}

	writeSectionHeader(&builder, "Body", color, colorCyan)
	if response.Body == "" {
		builder.WriteString("(empty)\n")
	} else {
		builder.WriteString(style(response.Body, color, colorMagenta))
		builder.WriteString("\n")
	}

	if len(response.Warnings) > 0 {
		writeSectionHeader(&builder, "Warnings", color, colorCyan)
		for _, warning := range response.Warnings {
			builder.WriteString(style("- "+warning, color, colorRed))
			builder.WriteString("\n")
		}
	}

	if response.RequestLine != "" {
		writeSectionHeader(&builder, "Request", color, colorCyan)
		builder.WriteString(style(response.RequestLine, color, colorGreen))
		builder.WriteString("\n")
		if len(response.RequestHeaders) > 0 {
			for _, header := range response.RequestHeaders {
				builder.WriteString(fmt.Sprintf("%s: %s\n", style(header.Name, color, colorYellow), header.Value))
			}
		}
	}

	return builder.String()
}

func looksVerbose(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "HTTP/") {
			return false
		}
		if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "> ") || strings.HasPrefix(trimmed, "< ") {
			return true
		}
	}

	return false
}

func parseVerbose(lines []string) Response {
	response := Response{Mode: "verbose"}
	inResponseHeaders := false
	inResponseBody := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inResponseHeaders {
				inResponseHeaders = false
				inResponseBody = true
			}
			continue
		}

		if trimmed == "<" {
			if inResponseHeaders {
				inResponseHeaders = false
				inResponseBody = true
			}
			continue
		}

		if strings.HasPrefix(trimmed, "* ") {
			continue
		}

		if strings.HasPrefix(trimmed, "> ") {
			payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "> "))
			if payload == "" {
				continue
			}
			if response.RequestLine == "" && strings.Contains(payload, " HTTP/") {
				response.RequestLine = payload
				continue
			}
			if name, value, ok := strings.Cut(payload, ":"); ok {
				response.RequestHeaders = append(response.RequestHeaders, Header{Name: strings.TrimSpace(name), Value: strings.TrimSpace(value)})
			}
			continue
		}

		if strings.HasPrefix(trimmed, "< ") {
			payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "< "))
			if payload == "" {
				inResponseHeaders = false
				inResponseBody = true
				continue
			}
			if strings.HasPrefix(payload, "HTTP/") {
				response.StatusLine = payload
				inResponseHeaders = true
				inResponseBody = false
				continue
			}
			if inResponseHeaders {
				if name, value, ok := strings.Cut(payload, ":"); ok {
					response.Headers = append(response.Headers, Header{Name: strings.TrimSpace(name), Value: strings.TrimSpace(value)})
					continue
				}
			}
		}

		if inResponseBody {
			if response.Body != "" {
				response.Body += "\n"
			}
			response.Body += line
		}
	}

	if response.StatusLine == "" {
		response.Warnings = append(response.Warnings, "missing response status line in verbose output")
	}
	if response.Body == "" && len(response.Headers) == 0 {
		response.Warnings = append(response.Warnings, "unable to detect response sections")
	}

	return response
}

func writeSectionHeader(builder *strings.Builder, title string, color bool, code string) {
	builder.WriteString(style(title, color, code))
	builder.WriteString("\n")
	builder.WriteString(style(strings.Repeat("=", len(title)), color, code))
	builder.WriteString("\n")
}

func style(value string, color bool, code string) string {
	if !color {
		return value
	}
	return code + value + colorReset
}

func isJSONContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json")
}

const (
	colorReset   = "\033[0m"
	colorCyan    = "\033[36m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorRed     = "\033[31m"
)
