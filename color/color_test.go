package color

import (
	"os"
	"testing"
)

func TestDisable(t *testing.T) {
	// Reset state before test
	globalDisable = false
	os.Unsetenv("NO_COLOR")

	if IsDisabled() {
		t.Errorf("Expected IsDisabled() to be false initially")
	}

	Disable()
	if !IsDisabled() {
		t.Errorf("Expected IsDisabled() to be true after calling Disable()")
	}

	// Reset state
	globalDisable = false
}

func TestNoColorEnv(t *testing.T) {
	// Reset state before test
	globalDisable = false
	os.Unsetenv("NO_COLOR")

	if IsDisabled() {
		t.Errorf("Expected IsDisabled() to be false without NO_COLOR")
	}

	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	if !IsDisabled() {
		t.Errorf("Expected IsDisabled() to be true when NO_COLOR is set")
	}
}

func TestWrap(t *testing.T) {
	// Reset state before test
	globalDisable = false
	os.Unsetenv("NO_COLOR")

	// Test when enabled and no global disable
	res := Wrap(true, Red, "test")
	if res != Red+"test"+Reset {
		t.Errorf("Wrap(true) failed. Got: %s", res)
	}

	// Test when disabled locally
	res = Wrap(false, Red, "test")
	if res != "test" {
		t.Errorf("Wrap(false) failed. Got: %s", res)
	}

	// Test when enabled locally but globally disabled via flag
	Disable()
	res = Wrap(true, Red, "test")
	if res != "test" {
		t.Errorf("Wrap(true) with Disable() failed. Got: %s", res)
	}
	globalDisable = false

	// Test when enabled locally but globally disabled via env
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	res = Wrap(true, Red, "test")
	if res != "test" {
		t.Errorf("Wrap(true) with NO_COLOR failed. Got: %s", res)
	}
}
