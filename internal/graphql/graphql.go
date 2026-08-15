package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kavix/kurl/client"
)

type Options struct {
	URL           string
	Query         string
	Variables     string
	Introspect    bool
	GenerateQuery string
	Headers       []string
	Verbose       bool
}

type Payload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

const IntrospectionQuery = `
query IntrospectionQuery {
  __schema {
    types {
      name
      kind
      fields {
        name
        type {
          name
          kind
          ofType {
            name
            kind
          }
        }
      }
    }
  }
}
`

// BuildPayload constructs the JSON payload for GraphQL requests
func BuildPayload(queryStr, variablesStr string, introspect bool) ([]byte, error) {
	query := queryStr
	if introspect {
		query = IntrospectionQuery
	}

	if query == "" {
		return nil, fmt.Errorf("GraphQL query is required (use --query or --introspect)")
	}

	payload := Payload{
		Query: query,
	}

	if variablesStr != "" {
		var vars map[string]interface{}
		if err := json.Unmarshal([]byte(variablesStr), &vars); err != nil {
			return nil, fmt.Errorf("invalid GraphQL variables JSON: %w", err)
		}
		payload.Variables = vars
	}

	return json.Marshal(payload)
}

// GenerateQueryForType creates a basic GraphQL query snippet for a type name
func GenerateQueryForType(typeName string) string {
	return fmt.Sprintf("query Get%s {\n  %s {\n    id\n    name\n  }\n}", typeName, typeName)
}

// ExecuteGraphQL sends a POST request with GraphQL JSON body to the target endpoint
func ExecuteGraphQL(opts Options) (*client.Result, error) {
	if opts.GenerateQuery != "" {
		generated := GenerateQueryForType(opts.GenerateQuery)
		fmt.Fprintf(os.Stdout, "Generated GraphQL Query for %s:\n%s\n\n", opts.GenerateQuery, generated)
		if opts.Query == "" {
			opts.Query = generated
		}
	}

	payloadBytes, err := BuildPayload(opts.Query, opts.Variables, opts.Introspect)
	if err != nil {
		return nil, err
	}

	headers := append([]string{"Content-Type: application/json"}, opts.Headers...)

	fetchOpts := client.Options{
		Method:  http.MethodPost,
		URL:     opts.URL,
		Data:    payloadBytes,
		Headers: headers,
		Verbose: opts.Verbose,
	}

	return client.Fetch(fetchOpts)
}
