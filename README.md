# kurl

`kurl` is a fast, colorized Go CLI for viewing HTTP API responses in a clean, human-readable format.

It focuses on the things you usually want first: status, timing, protocol, headers, and a readable body.

## Features

- `kurl [METHOD] <URL> [flags]` command shape.
- Automatic `https://` and `http://` fallback for bare hosts like `google.lk`.
- Concurrent scheme probing when no scheme is provided.
- Tuned HTTP client with keep-alives, HTTP/2, timeouts, and custom DNS dialing.
- Pretty JSON formatting with recursive color.
- Raw body mode for exact output.
- Verbose request/redirect details.
- Automatic color disabling for non-TTY output or `NO_COLOR`.

## Install

Build the binary:

```bash
go build -ldflags="-s -w" -o kurl .
```

Install to your PATH:

```bash
make install
```

If you use the default install target, the binary is placed in `~/.local/bin`.

## Quick Start

Fetch a URL:

```bash
kurl https://api.genderize.io/?name=luc
```

Use a bare host:

```bash
kurl google.lk
```

POST JSON:

```bash
kurl POST https://api.example.com/users -d '{"name":"luc"}' -H "Authorization: Bearer token"
```

## Output

Typical output looks like this:

```text
┌─────────────────────────────────────────────┐
│  kurl · GET https://api.genderize.io/?name=luc │
└─────────────────────────────────────────────┘

	STATUS   200 OK
	TIME     142ms
	PROTO    HTTP/2.0

── HEADERS ───────────────────────────────────
	Content-Type       application/json; charset=utf-8
	X-Rate-Limit       1000

── BODY ──────────────────────────────────────
{
	"count": 736,
	"gender": "male",
	"name": "luc",
	"probability": 0.98
}
```

## Flags

| Flag | Description |
| --- | --- |
| `-X`, `--method` | HTTP method to use. Default: `GET` |
| `-d`, `--data` | Request body |
| `-H`, `--header` | Add a request header. Repeatable |
| `-t`, `--timeout` | Timeout in seconds. Default: `30` |
| `--no-color` | Disable ANSI color output |
| `--headers-only` | Show only response headers |
| `--body-only` | Show only response body |
| `--raw` | Print the body with no formatting |
| `-v`, `--verbose` | Show request info and redirect chain |
| `-o`, `--output` | Save the response body to a file |

## Examples

### Simple GET

```bash
kurl https://api.genderize.io/?name=luc
```

### Explicit HTTP method

```bash
kurl POST https://api.example.com/users
```

### JSON request body

```bash
kurl POST https://api.example.com/users -d '{"name":"luc"}' -H "Content-Type: application/json"
```

### Multiple headers

```bash
kurl GET https://api.example.com/me -H "Authorization: Bearer token" -H "X-Client: kurl"
```

### Verbose request details

```bash
kurl -v https://example.com
```

### JSON response formatting

```bash
kurl https://api.example.com/data
```

If the response is JSON, `kurl` pretty-prints it automatically.

### Save body to disk

```bash
kurl https://example.com/image.png -o image.png
```

### Headers only

```bash
kurl --headers-only https://example.com
```

### Body only

```bash
kurl --body-only https://example.com
```

### Raw output

```bash
kurl --raw https://example.com
```

### Disable color

```bash
kurl --no-color https://example.com
```

## Behavior

- If you pass a URL without a scheme, `kurl` tries `https://` and `http://` concurrently.
- The first successful response wins.
- JSON bodies are pretty-printed with 2-space indentation.
- Non-JSON text bodies are shown as-is.
- Binary content is replaced with `[Binary data - use -o to save]`.
- Empty bodies show `No body`.

## Redirects

With `--verbose`, redirect hops are shown in order so you can see where the request moved.

## Development

Run tests:

```bash
go test ./...
```

Build the optimized binary:

```bash
go build -ldflags="-s -w" -o kurl .
```

## Notes

- The default install target writes to `~/.local/bin`.
- `NO_COLOR` disables color automatically.
- Output is easier to read in a real terminal; redirecting to a file disables color detection.