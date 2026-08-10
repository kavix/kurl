package color

import (
	"bytes"
	"os"
	"testing"
)

func TestColorHelperFunctions(t *testing.T) {
	globalDisable = false
	os.Unsetenv("NO_COLOR")

	if Border(true, "test") != Dim+White+"test"+Reset {
		t.Errorf("Border failed")
	}
	if Title(true, "test") != Bold+Blue+"test"+Reset {
		t.Errorf("Title failed")
	}
	if Header(true, "test") != Bold+White+"test"+Reset {
		t.Errorf("Header failed")
	}
	if Key(true, "test") != Cyan+"test"+Reset {
		t.Errorf("Key failed")
	}
	if String(true, "test") != Green+"test"+Reset {
		t.Errorf("String failed")
	}
	if Number(true, "123") != Yellow+"123"+Reset {
		t.Errorf("Number failed")
	}
	if Bool(true, "true") != Magenta+"true"+Reset {
		t.Errorf("Bool failed")
	}
	if Null(true, "null") != Red+"null"+Reset {
		t.Errorf("Null failed")
	}
	if ErrorText(true, "error") != Red+"error"+Reset {
		t.Errorf("ErrorText failed")
	}
}

func TestStatusColoring(t *testing.T) {
	globalDisable = false
	os.Unsetenv("NO_COLOR")

	if Status(true, 200, "200 OK") != Bold+Green+"200 OK"+Reset {
		t.Errorf("Status 200 failed")
	}
	if Status(true, 301, "301 Moved") != Bold+Yellow+"301 Moved"+Reset {
		t.Errorf("Status 301 failed")
	}
	if Status(true, 404, "404 Not Found") != Bold+Red+"404 Not Found"+Reset {
		t.Errorf("Status 404 failed")
	}
	if Status(true, 500, "500 Error") != Bold+Red+"500 Error"+Reset {
		t.Errorf("Status 500 failed")
	}
}

func TestAutoEnabledAndFprintf(t *testing.T) {
	globalDisable = false
	os.Unsetenv("NO_COLOR")

	var buf bytes.Buffer
	if AutoEnabled(&buf) {
		t.Errorf("Expected AutoEnabled to return false for bytes.Buffer")
	}

	n, err := Fprintf(&buf, "hello %s", "world")
	if err != nil || n != 11 || buf.String() != "hello world" {
		t.Errorf("Fprintf failed")
	}
}
