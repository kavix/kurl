package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"brew-terminal-curl/color"
)

func PrettyJSON(w io.Writer, r io.Reader, enabled bool) (int64, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var count int64
	if err := writeJSONValue(w, dec, enabled, 0, &count); err != nil {
		return count, err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return count, err
	}
	count++
	return count, nil
}

func writeJSONValue(w io.Writer, dec *json.Decoder, enabled bool, depth int, count *int64) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	return writeJSONToken(w, dec, tok, enabled, depth, count)
}

func writeJSONToken(w io.Writer, dec *json.Decoder, tok json.Token, enabled bool, depth int, count *int64) error {
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			if _, err := fmt.Fprint(w, "{"); err != nil { return err }
			if _, err := fmt.Fprint(w, "\n"); err != nil { return err }
			*count += 2
			first := true
			for dec.More() {
				if !first {
					if _, err := fmt.Fprint(w, ",\n"); err != nil { return err }
					*count += 2
				}
				first = false
				if _, err := fmt.Fprint(w, strings.Repeat("  ", depth+1)); err != nil { return err }
				*count += int64(len(strings.Repeat("  ", depth+1)))
				keyTok, err := dec.Token()
				if err != nil { return err }
				key, ok := keyTok.(string)
				if !ok { return fmt.Errorf("invalid json key") }
				if _, err := fmt.Fprintf(w, "%s: ", color.Key(enabled, fmt.Sprintf("\"%s\"", key))); err != nil { return err }
				*count += int64(len(key))
				if err := writeJSONValue(w, dec, enabled, depth+1, count); err != nil { return err }
			}
			if _, err := dec.Token(); err != nil { return err }
			if _, err := fmt.Fprint(w, "\n"+strings.Repeat("  ", depth)+"}"); err != nil { return err }
			*count += int64(len(strings.Repeat("  ", depth))) + 1
			return nil
		case '[':
			if _, err := fmt.Fprint(w, "["); err != nil { return err }
			if _, err := fmt.Fprint(w, "\n"); err != nil { return err }
			*count += 2
			first := true
			for dec.More() {
				if !first {
					if _, err := fmt.Fprint(w, ",\n"); err != nil { return err }
					*count += 2
				}
				first = false
				if _, err := fmt.Fprint(w, strings.Repeat("  ", depth+1)); err != nil { return err }
				if err := writeJSONValue(w, dec, enabled, depth+1, count); err != nil { return err }
			}
			if _, err := dec.Token(); err != nil { return err }
			if _, err := fmt.Fprint(w, "\n"+strings.Repeat("  ", depth)+"]"); err != nil { return err }
			return nil
		default:
			return fmt.Errorf("unexpected delimiter %q", v)
		}
	case string:
		if _, err := fmt.Fprint(w, color.String(enabled, fmt.Sprintf("\"%s\"", v))); err != nil { return err }
	case json.Number:
		if _, err := fmt.Fprint(w, color.Number(enabled, v.String())); err != nil { return err }
	case bool:
		if _, err := fmt.Fprint(w, color.Bool(enabled, fmt.Sprintf("%t", v))); err != nil { return err }
	case nil:
		if _, err := fmt.Fprint(w, color.Null(enabled, "null")); err != nil { return err }
	default:
		if _, err := fmt.Fprint(w, fmt.Sprint(v)); err != nil { return err }
	}
	return nil
}