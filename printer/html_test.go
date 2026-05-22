package printer

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrettyHTMLInlineHandling(t *testing.T) {
	input := `<div>Hello <b>world</b>!</div>`
	var buf bytes.Buffer
	_, err := PrettyHTML(&buf, strings.NewReader(input), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	// Should be printed inline on a single line because all children are inline
	expected := "<div>Hello <b>world</b>!</div>\n"
	if result != expected {
		t.Fatalf("expected:\n%q\ngot:\n%q", expected, result)
	}
}

func TestPrettyHTMLBlockFormatting(t *testing.T) {
	input := `<html><body><div class="main"><h1>Title</h1><p>Para</p></div></body></html>`
	var buf bytes.Buffer
	_, err := PrettyHTML(&buf, strings.NewReader(input), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	// Should format blocks with clean indentation
	expectedLines := []string{
		"<html>",
		"  <body>",
		`    <div class="main">`,
		"      <h1>Title</h1>",
		"      <p>Para</p>",
		"    </div>",
		"  </body>",
		"</html>",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Fatalf("expected:\n%q\ngot:\n%q", expected, result)
	}
}

func TestPrettyHTMLVoidTagsAndComments(t *testing.T) {
	input := `<div><!-- comment -->Line 1<br><img src="pic.png"></div>`
	var buf bytes.Buffer
	_, err := PrettyHTML(&buf, strings.NewReader(input), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	// Since <div> has multiple items but br and img are inline elements, the whole <div> has only inline children recursively!
	// So it should format as a single inline line.
	expected := "<div><!-- comment -->Line 1<br /><img src=\"pic.png\" /></div>\n"
	if result != expected {
		t.Fatalf("expected:\n%q\ngot:\n%q", expected, result)
	}
}

func TestPrettyHTMLColorization(t *testing.T) {
	input := `<div class="container">Content</div>`
	var buf bytes.Buffer
	_, err := PrettyHTML(&buf, strings.NewReader(input), true) // enabled = true
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	// Verify that color escape sequences (like \033) are present in the output
	if !strings.Contains(result, "\033[") {
		t.Fatalf("expected ANSI color codes in output, got: %q", result)
	}

	// Verify that tag name "div" and attribute "class" are present
	if !strings.Contains(result, "div") || !strings.Contains(result, "class") {
		t.Fatalf("expected formatted text to contain tag and attribute names, got: %q", result)
	}
}
