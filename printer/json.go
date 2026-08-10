package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"brew-terminal-curl/color"
)

var indentCache = func() [32][]byte {
	var cache [32][]byte
	for i := 0; i < 32; i++ {
		cache[i] = []byte(strings.Repeat("  ", i))
	}
	return cache
}()

func getIndent(depth int) []byte {
	if depth < 32 {
		return indentCache[depth]
	}
	return []byte(strings.Repeat("  ", depth))
}

func PrettyJSON(w io.Writer, r io.Reader, enabled bool) (int64, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var count int64
	if err := writeJSONValue(w, dec, enabled, 0, &count); err != nil {
		return count, err
	}
	if _, err := w.Write([]byte("\n")); err != nil {
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
			if _, err := w.Write([]byte("{\n")); err != nil {
				return err
			}
			*count += 2
			first := true
			for dec.More() {
				if !first {
					if _, err := w.Write([]byte(",\n")); err != nil {
						return err
					}
					*count += 2
				}
				first = false
				indent := getIndent(depth + 1)
				if _, err := w.Write(indent); err != nil {
					return err
				}
				*count += int64(len(indent))
				keyTok, err := dec.Token()
				if err != nil {
					return err
				}
				key, ok := keyTok.(string)
				if !ok {
					return fmt.Errorf("invalid json key")
				}

				keyFormatted := color.Key(enabled, `"`+key+`"`) + ": "
				if _, err := io.WriteString(w, keyFormatted); err != nil {
					return err
				}
				*count += int64(len(key))
				if err := writeJSONValue(w, dec, enabled, depth+1, count); err != nil {
					return err
				}
			}
			if _, err := dec.Token(); err != nil {
				return err
			}
			closeStr := "\n" + string(getIndent(depth)) + "}"
			if _, err := io.WriteString(w, closeStr); err != nil {
				return err
			}
			*count += int64(len(closeStr))
			return nil
		case '[':
			if _, err := w.Write([]byte("[\n")); err != nil {
				return err
			}
			*count += 2
			first := true
			for dec.More() {
				if !first {
					if _, err := w.Write([]byte(",\n")); err != nil {
						return err
					}
					*count += 2
				}
				first = false
				indent := getIndent(depth + 1)
				if _, err := w.Write(indent); err != nil {
					return err
				}
				*count += int64(len(indent))
				if err := writeJSONValue(w, dec, enabled, depth+1, count); err != nil {
					return err
				}
			}
			if _, err := dec.Token(); err != nil {
				return err
			}
			closeStr := "\n" + string(getIndent(depth)) + "]"
			if _, err := io.WriteString(w, closeStr); err != nil {
				return err
			}
			*count += int64(len(closeStr))
			return nil
		default:
			return fmt.Errorf("unexpected delimiter %q", v)
		}
	case string:
		strFormatted := color.String(enabled, `"`+v+`"`)
		if _, err := io.WriteString(w, strFormatted); err != nil {
			return err
		}
	case json.Number:
		numFormatted := color.Number(enabled, v.String())
		if _, err := io.WriteString(w, numFormatted); err != nil {
			return err
		}
	case bool:
		boolStr := strconv.FormatBool(v)
		boolFormatted := color.Bool(enabled, boolStr)
		if _, err := io.WriteString(w, boolFormatted); err != nil {
			return err
		}
	case nil:
		nullFormatted := color.Null(enabled, "null")
		if _, err := io.WriteString(w, nullFormatted); err != nil {
			return err
		}
	default:
		if _, err := io.WriteString(w, fmt.Sprint(v)); err != nil {
			return err
		}
	}
	return nil
}
