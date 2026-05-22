# 🚀 kurl

`kurl` is a lightning-fast, beautifully colorized Go command-line tool designed to fetch HTTP APIs and render responses in a clean, human-readable format. 

Forget raw, unformatted, monochrome terminal dumps. `kurl` focuses on the information you actually want to see first—**status codes, request times, protocols, headers, and perfectly formatted, syntax-highlighted bodies.**

---

## ✨ Features

*   **⚡ Concurrency-Powered Probing**: Pass a raw domain (e.g. `google.com`), and `kurl` will query `https://` and `http://` in parallel, automatically resolving and serving the fastest successful candidate.
*   **🏎️ Concurrent DNS Racing Resolver**: Multi-threaded DNS racing queries both the default system DNS and Cloudflare's `1.1.1.1` in parallel. Bypasses VPN bottlenecks, slow local resolvers, and eliminates DNS hang latencies entirely.
*   **🎨 Token-by-Token JSON Formatter**: Parses JSON response bodies on the fly, rendering them with strict indentation and harmonized syntax-highlighting terminal colors.
*   **🌳 Smart HTML Pretty-Printer**: Leverages a full HTML5 compliant DOM parser to format structure with `2-space` indentations, collapse inline element nodes (`<b>`, `<i>`, `<a>`, `<span>`) onto single vertical lines to avoid line bloat, and colorize tags, attributes, values, and comments.
*   **🛡️ CDN & Anti-Bot Protection Bypass**: Automatically injects standard modern browser headers (`User-Agent` and `Accept`) to prevent anti-bot blocking layers (like Cloudflare or Akamai) from rejecting your CLI requests.
*   **🔌 Smart TTY Output Switching**: Automatically detects if the output stream is redirected to a file, pipe, or script, and silently strips ANSI colors. Respects the standard `NO_COLOR` environment variable.
*   **🔍 Verbose Request Chains**: Trace full redirect hops (`301`, `302`, `307`, `308`) and inspect absolute request-to-response headers instantly using the `-v` flag.

---

## 📦 Installation

### 1. via Homebrew (Recommended)
`kurl` is published to your personal Homebrew Tap for seamless cross-platform installation:

```bash
# Add the tap and install kurl
brew tap kavix/tap
brew install kurl
```

To update `kurl` to the latest version in the future:
```bash
brew update
brew upgrade kurl
```

### 2. From Source
If you have a Go toolchain installed (`Go 1.22` or newer):

```bash
# Clone the repository
git clone https://github.com/kavix/kurl.git
cd kurl

# Compile and place the binary into ~/.local/bin
make install
```

The default Makefile target compiles an optimized binary with stripped debugging symbols and places it under `~/.local/bin`. Ensure this directory is in your shell's `PATH`.

---

## 🚀 Quick Start

Fetch any API or web page with a simple command:

```bash
# Fetch and format a JSON API
kurl https://api.genderize.io/?name=luc

# Fetch a webpage with automatic scheme probing and smart HTML rendering
kurl news.lk

# POST JSON payload with custom headers
kurl POST https://api.example.com/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-token" \
  -d '{"name": "Alice", "role": "admin"}'
```

---

## 📥 Command Usage & Flags

```text
kurl [METHOD] <URL> [flags]
```

### Flags

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

## 🎨 Visual Formatting System

`kurl` automatically styles payloads to give you an exceptionally readable presentation:

### 1. Title Bar & Response Metadata
Every request is printed with an organized header block detailing connection and network metrics:
```text
┌─────────────────────────────────────────────┐
│  kurl · GET https://api.genderize.io/?name=luc │
└─────────────────────────────────────────────┘
  STATUS   200 OK
  TIME     120ms
  PROTO    HTTP/2.0
```

### 2. JSON Format Style
Parses, structures, and highlights properties:
*   **Braces/Brackets**: Gray/Dim
*   **Keys**: Cyan
*   **Strings**: Green
*   **Numbers & Booleans**: Yellow
*   **Nulls**: Bold Red

### 3. HTML Format Style
Uses a full tree-parser to restructure documents with beautiful indentations:
*   **Tags & Bracket Closures**: Cyan & Dim Gray
*   **Attributes Keys**: Yellow
*   **Attribute Values**: Green
*   **DOCTYPE Definitions**: Bold Magenta
*   **HTML Comments**: Dim Gray / Italic

---

## 🔧 Developer & Contribution Guide

### Running Tests
`kurl` has unit test suites covering the DNS concurrent racing system, JSON tokenization, HTML collapsing, and CLI parser rules. Run them with:

```bash
go test -v ./...
```

### Compiling Production Binaries
To build a highly optimized binary locally with stripped debug symbols and optimized size:

```bash
make build
```

### Snapshot Compilation (Multi-Platform)
To test cross-compilation on your machine for macOS, Linux, and Windows:

```bash
make release-local
```

### Triggering a Production Release
Releases are fully automated via GitHub Actions ([release.yml](file:///.github/workflows/release.yml)):
1. Create a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   ```
2. Push it to GitHub:
   ```bash
   git push origin v1.0.0
   ```
3. The runner will compile the binaries, upload them to GitHub Releases, and automatically commit the new Formula to your [kavix/homebrew-tap](https://github.com/kavix/homebrew-tap) repository.

---

## 📄 License

`kurl` is open-source software licensed under the **MIT License**.