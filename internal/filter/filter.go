package filter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ApplyFilter evaluates a query string against raw JSON data.
// Supported queries:
// - `.field` or `.field.subfield`
// - `.array[0]` or `.array[1].name`
// - `.array[]` or `.array[].name` (flattens/projects array elements)
func ApplyFilter(jsonData []byte, query string) ([]byte, error) {
	query = strings.TrimSpace(query)
	if query == "" || query == "." {
		return jsonData, nil
	}

	var root interface{}
	if err := json.Unmarshal(jsonData, &root); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	res, err := evalPath(root, query)
	if err != nil {
		return nil, err
	}

	return json.Marshal(res)
}

func evalPath(val interface{}, query string) (interface{}, error) {
	if query == "" || query == "." {
		return val, nil
	}

	tokens, err := parseTokens(query)
	if err != nil {
		return nil, err
	}

	curr := val
	for _, tok := range tokens {
		if curr == nil {
			return nil, nil
		}
		next, err := stepToken(curr, tok)
		if err != nil {
			return nil, err
		}
		curr = next
	}

	return curr, nil
}

type token struct {
	kind   string // "field", "index", "all"
	val    string
	idx    int
}

func parseTokens(query string) ([]token, error) {
	if !strings.HasPrefix(query, ".") {
		query = "." + query
	}

	var tokens []token
	parts := strings.Split(query, ".")

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array indexing like `users[0]` or `users[]` or `[0]`
		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			openIdx := strings.Index(part, "[")
			closeIdx := strings.Index(part, "]")
			
			fieldName := part[:openIdx]
			bracketVal := part[openIdx+1 : closeIdx]

			if fieldName != "" {
				tokens = append(tokens, token{kind: "field", val: fieldName})
			}

			if bracketVal == "" {
				tokens = append(tokens, token{kind: "all"})
			} else {
				idx, err := strconv.Atoi(bracketVal)
				if err != nil {
					return nil, fmt.Errorf("invalid array index %q in query", bracketVal)
				}
				tokens = append(tokens, token{kind: "index", idx: idx})
			}
		} else {
			tokens = append(tokens, token{kind: "field", val: part})
		}
	}

	return tokens, nil
}

func stepToken(curr interface{}, tok token) (interface{}, error) {
	switch tok.kind {
	case "field":
		obj, ok := curr.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access field %q on non-object %v", tok.val, reflect.TypeOf(curr))
		}
		return obj[tok.val], nil

	case "index":
		arr, ok := curr.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot index non-array")
		}
		if tok.idx < 0 || tok.idx >= len(arr) {
			return nil, fmt.Errorf("array index %d out of bounds (len %d)", tok.idx, len(arr))
		}
		return arr[tok.idx], nil

	case "all":
		arr, ok := curr.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot iterate all elements on non-array")
		}
		return arr, nil

	default:
		return nil, fmt.Errorf("unknown token kind %q", tok.kind)
	}
}

// FilterKeys extracts specific keys from a JSON object or array of objects.
func FilterKeys(jsonData []byte, keysCSV string) ([]byte, error) {
	rawKeys := strings.Split(keysCSV, ",")
	var keys []string
	for _, k := range rawKeys {
		if trimmed := strings.TrimSpace(k); trimmed != "" {
			keys = append(keys, trimmed)
		}
	}

	var root interface{}
	if err := json.Unmarshal(jsonData, &root); err != nil {
		return nil, err
	}

	filtered := filterKeysValue(root, keys)
	return json.Marshal(filtered)
}

func filterKeysValue(val interface{}, keys []string) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		res := make(map[string]interface{})
		for _, k := range keys {
			if item, exists := v[k]; exists {
				res[k] = item
			}
		}
		return res
	case []interface{}:
		var res []interface{}
		for _, elem := range v {
			res = append(res, filterKeysValue(elem, keys))
		}
		return res
	default:
		return val
	}
}

// FlattenArray flattens nested arrays in JSON data.
func FlattenArray(jsonData []byte) ([]byte, error) {
	var root interface{}
	if err := json.Unmarshal(jsonData, &root); err != nil {
		return nil, err
	}

	arr, ok := root.([]interface{})
	if !ok {
		return jsonData, nil
	}

	var result []interface{}
	var flatten func(item interface{})
	flatten = func(item interface{}) {
		if innerArr, isArr := item.([]interface{}); isArr {
			for _, inner := range innerArr {
				flatten(inner)
			}
		} else {
			result = append(result, item)
		}
	}

	for _, item := range arr {
		flatten(item)
	}

	return json.Marshal(result)
}
