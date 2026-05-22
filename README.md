# kurl

`kurl` is a fast, colorized Go CLI for viewing HTTP API responses in a clean, human-readable format.

It focuses on the things you usually want first: status, timing, protocol, headers, and a readable body.

## Features

- `kurl [METHOD] <URL> [flags]` command shape.
- **Smart HTML Formatter**: Prettifies, auto-indents (2-space), and syntax-highlights HTML responses with smart collapsing for inline element trees.
- **Parallel DNS Racing**: Concurrently queries public Cloudflare DNS and host system DNS, racing them to connect using the fastest available resolver.
- **CDN Protection Bypass**: Injects clean browser-like default headers to prevent server-side bot-blocking filters.
- **Concurrent Protocol Probing**: Probes `https://` and `http://` in parallel when no scheme is provided for quick automatic fallback.
- **Pretty JSON Formatter**: Formats and styles JSON body responses token-by-token using recursive terminal colors.
- **Tuned Keep-Alives & HTTP/2**: Forces HTTP/2 where available, with optimized TCP connect and TLS handshake parameters.
- **Automatic TTY Color detection**: Smart disables terminal color sequences when outputting to files, pipes, or when `NO_COLOR` is present.
- **Raw & Verbose Output modes**: Toggleable raw bodies and full request/redirect chain diagnostics.

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

## Release & Publishing

`kurl` is ready to compile and publish across all platforms (macOS, Linux, Windows) with automated GitHub Releases and Homebrew Tap support using **GoReleaser**.

### 1. Local Pre-check & Snapshot Builds
To test compile and bundle the binaries locally without pushing or publishing:

```bash
# Make sure goreleaser is installed
brew install goreleaser/tap/goreleaser

# Run snapshot build via Makefile
make release-local
```
This builds multi-platform binaries in the `dist/` directory.

### 2. Automating Releases with GitHub Actions
We've set up a pre-configured CI/CD pipeline in `.github/workflows/release.yml`. When you're ready to publish:

1. Draft a Git version tag:
   ```bash
   git tag -a v1.0.0 -m "First release"
   ```
2. Push the tag to GitHub:
   ```bash
   git push origin v1.0.0
   ```
3. The GitHub Action will automatically:
   - Compile binaries for all target OS and architectures.
   - Build `.tar.gz` and `.zip` release archives.
   - Calculate hash checksums.
   - Publish the artifacts to a new GitHub Release.
   - Automatically push the formula file to your Homebrew Tap (`homebrew-tap`).

## Notes

- The default install target writes to `~/.local/bin`.
- `NO_COLOR` disables color automatically.
- Output is easier to read in a real terminal; redirecting to a file disables color detection.