package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type environmentConfig struct {
	BaseURL string   `json:"base_url"`
	Headers []string `json:"headers,omitempty"`
}

func loadEnvironment(name string) (environmentConfig, error) {
	var config environmentConfig
	home, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("unable to find home directory: %w", err)
	}

	dir := home + "/.kurl"
	filePath := dir + "/environments.json"

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create default environments template
		if err := os.MkdirAll(dir, 0755); err != nil {
			return config, fmt.Errorf("unable to create config directory: %w", err)
		}

		defaultData := `{
  "dev": {
    "base_url": "http://localhost:8080/v1",
    "headers": [
      "X-Environment: development",
      "Authorization: Bearer dev-token"
    ]
  },
  "prod": {
    "base_url": "https://api.example.com/v1",
    "headers": [
      "X-Environment: production",
      "Authorization: Bearer prod-token"
    ]
  }
}`
		if err := os.WriteFile(filePath, []byte(defaultData), 0644); err != nil {
			return config, fmt.Errorf("unable to write default environments config: %w", err)
		}
		fmt.Printf("ℹ️ Created default environments template at %s\n", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("unable to read environments config: %w", err)
	}

	var envs map[string]environmentConfig
	if err := json.Unmarshal(data, &envs); err != nil {
		return config, fmt.Errorf("unable to parse environments config: %w", err)
	}

	envConfig, exists := envs[name]
	if !exists {
		return config, fmt.Errorf("environment %q not found in %s", name, filePath)
	}

	return envConfig, nil
}

func applyEnvironment(opts *cliOptions) error {
	if opts.env == "" {
		return nil
	}

	envConfig, err := loadEnvironment(opts.env)
	if err != nil {
		return err
	}

	// 1. Process URL
	hasScheme := strings.HasPrefix(opts.url, "http://") ||
		strings.HasPrefix(opts.url, "https://") ||
		strings.HasPrefix(opts.url, "ws://") ||
		strings.HasPrefix(opts.url, "wss://")

	if !hasScheme {
		opts.url = joinURL(envConfig.BaseURL, opts.url)
	}

	// 2. Process headers (merge)
	opts.headers = mergeHeaders(envConfig.Headers, opts.headers)

	return nil
}

func joinURL(baseURL, path string) string {
	if baseURL == "" {
		return path
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return baseURL
	}
	return baseURL + "/" + path
}

func mergeHeaders(profileHeaders []string, cliHeaders []string) []string {
	result := append([]string(nil), profileHeaders...)

	for _, cliHeader := range cliHeaders {
		cliName, _, ok := strings.Cut(cliHeader, ":")
		if !ok {
			result = append(result, cliHeader)
			continue
		}
		cliNameClean := strings.ToLower(strings.TrimSpace(cliName))

		// Check if this header already exists in result
		found := false
		for idx, resHeader := range result {
			resName, _, ok := strings.Cut(resHeader, ":")
			if ok && strings.ToLower(strings.TrimSpace(resName)) == cliNameClean {
				// Replace it!
				result[idx] = cliHeader
				found = true
				break
			}
		}
		if !found {
			result = append(result, cliHeader)
		}
	}
	return result
}
