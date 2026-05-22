# 🚀 kurl Usage Guide

`kurl` is designed to be simple, fast, and drop-in compatible with standard workflows, but with enhanced capabilities out-of-the-box. This guide will walk you through common use cases, flags, and troubleshooting tips.

## Basic Usage

Fetch any API or web page with a simple command:

```bash
# Fetch and format a JSON API. kurl auto-formats JSON responses beautifully.
kurl https://api.genderize.io/?name=luc

# Fetch a webpage with automatic scheme probing (http:// vs https://) and smart HTML rendering
kurl news.lk
```

## Advanced Requests

### Sending Data (POST/PUT)
You can easily send JSON or raw data payloads using the `-d` (or `--data`) flag. `kurl` automatically infers `POST` if you pass data, but it's best practice to specify the method explicitly.

```bash
kurl POST https://api.example.com/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-token" \
  -d '{"name": "Alice", "role": "admin"}'
```

### Custom Headers
Use `-H` (or `--header`) to attach headers. This can be repeated for multiple headers.
Note: `kurl` automatically injects standard modern browser headers (`User-Agent` and `Accept`) by default to bypass basic CDN/bot protections. If you specify your own `User-Agent`, it will override the default.

```bash
kurl GET https://api.github.com/users/kavix \
  -H "Accept: application/vnd.github.v3+json"
```

## Output Control & Tracing

### Verbose Mode
Trace full redirect hops (`301`, `302`, `307`, `308`) and inspect absolute request-to-response headers instantly using the `-v` flag.

```bash
kurl -v google.com
```

### Body or Headers Only
Sometimes you only want specific parts of the payload for scripting or piping.

```bash
# Print only the response body
kurl --body-only https://api.ipify.org?format=json

# Print only the response headers
kurl --headers-only https://example.com
```

### Disabling Formatters
If you want raw output without syntax highlighting or HTML collapsing (e.g., when saving to a file), use `--raw`.

```bash
kurl --raw -o response.html https://example.com
```

*Note: `kurl` has smart TTY output switching. If you pipe the output to another command or file, it will silently strip ANSI colors. You can also enforce this by setting `NO_COLOR=1` in your environment or using `--no-color`.*

## Command-Line Flags Reference

| Flag | Short | Description | Default |
| :--- | :--- | :--- | :--- |
| `--method` | `-X` | The HTTP request method | `GET` |
| `--data` | `-d` | Request body payload (JSON, raw text, etc.) | *None* |
| `--header` | `-H` | Custom request header (repeatable) | *None* |
| `--timeout` | `-t` | Max connection and transfer timeout in seconds | `30` |
| `--no-color` | *None* | Disables terminal ANSI colors | `false` |
| `--headers-only`| *None* | Excludes body output and prints headers only | `false` |
| `--body-only` | *None* | Excludes headers/stats and prints body only | `false` |
| `--raw` | *None* | Bypasses all pretty-printing engines | `false` |
| `--verbose` | `-v` | Prints request headers and complete redirect trace | `false` |
| `--output` | `-o` | Saves response body directly to a local file | *None* |
| `--version` | `-V` | Prints CLI build version, git commit, and date | `false` |

---

## Troubleshooting

### `zsh: no matches found` error

If you use **Zsh** and run a command with an unquoted URL containing question marks (`?`), ampersands (`&`), or equal signs (`=`), you may get a `zsh: no matches found` error. This is because Zsh attempts to expand these characters as wildcard globs.

**Solution 1 (Temporary):** Wrap your URLs in quotes:
```bash
kurl "https://api.example.com/?search=query"
```

**Solution 2 (Permanent & Recommended):** Add an alias to your `~/.zshrc` file that tells Zsh to automatically disable glob expansion specifically for the `kurl` command:
```bash
# Add this line to your ~/.zshrc
alias kurl="noglob kurl"
```
After saving, run `source ~/.zshrc`. You can now pass raw URLs to `kurl` without ever needing quotes again!
