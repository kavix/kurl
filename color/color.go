package color

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

func AutoEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

func Wrap(enabled bool, code string, value string) string {
	if !enabled {
		return value
	}
	return code + value + Reset
}

func Border(enabled bool, value string) string { return Wrap(enabled, Dim+White, value) }
func Title(enabled bool, value string) string  { return Wrap(enabled, Bold+Blue, value) }
func Header(enabled bool, value string) string { return Wrap(enabled, Bold+White, value) }
func Status(enabled bool, statusCode int, value string) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return Wrap(enabled, Bold+Green, value)
	case statusCode >= 300 && statusCode < 400:
		return Wrap(enabled, Bold+Yellow, value)
	default:
		return Wrap(enabled, Bold+Red, value)
	}
}
func Key(enabled bool, value string) string    { return Wrap(enabled, Cyan, value) }
func String(enabled bool, value string) string { return Wrap(enabled, Green, value) }
func Number(enabled bool, value string) string { return Wrap(enabled, Yellow, value) }
func Bool(enabled bool, value string) string   { return Wrap(enabled, Magenta, value) }
func Null(enabled bool, value string) string   { return Wrap(enabled, Red, value) }

func ErrorText(enabled bool, value string) string { return Wrap(enabled, Red, value) }
func Box(enabled bool, lines ...string) string    { return strings.Join(lines, "\n") }

func Fprintf(w io.Writer, format string, args ...any) (int, error) {
	return fmt.Fprintf(w, format, args...)
}
