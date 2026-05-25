# 🚀 kurl

<p align="center">
  <i>A lightning-fast, beautifully colorized Go command-line tool designed to fetch HTTP APIs and render responses in a clean, human-readable format.</i>
</p>

---

Forget raw, unformatted, monochrome terminal dumps. `kurl` focuses on the information you actually want to see first—**status codes, request times, protocols, headers, and perfectly formatted, syntax-highlighted bodies.**

## ✨ Why kurl?

*   **⚡ Concurrency-Powered Probing**: Pass a raw domain (e.g. `google.com`), and `kurl` queries `https://` and `http://` in parallel, serving the fastest successful candidate.
*   **🏎️ Concurrent DNS Racing**: Multi-threaded DNS racing queries default system DNS and Cloudflare's `1.1.1.1` in parallel to bypass VPN bottlenecks and eliminate DNS hang latencies entirely.
*   **🎨 Token-by-Token Formatter**: Parses JSON on the fly, rendering with strict indentation and harmonized syntax-highlighting.
*   **🌳 Smart HTML Pretty-Printer**: Leverages an HTML5 DOM parser to format structure cleanly and collapse inline element nodes to avoid line bloat.
*   **🛡️ Anti-Bot Bypass**: Automatically injects standard modern browser headers to prevent anti-bot blocking layers from rejecting your CLI requests.
*   **💾 Request Replays**: Save request profiles locally (like a terminal-native Postman) and replay them with option overrides.
*   **💬 Interactive WebSockets**: Connect to WebSocket endpoints with real-time duplex text frames and colorized formatted messages.
*   **🌐 Environment Profiles**: Switch base URLs and auth headers dynamically between dev/staging/prod environments using local configuration profiles.

## 📦 Installation

### macOS / Linux (Homebrew)
`kurl` is published to a Homebrew Tap for seamless cross-platform installation:

```bash
brew tap kavix/tap
brew install kurl
```

### From Source (Go 1.22+)
```bash
git clone https://github.com/kavix/kurl.git
cd kurl
make install
```
*Note: Make sure `~/.local/bin` is in your `PATH`.*

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

# Save a request configuration locally
kurl save github-api GET https://api.github.com/users/kavix -H "Accept: application/json"

# Replay the request with optional parameter overrides
kurl run github-api -v

# Open an interactive WebSocket session with colorized frames
kurl ws://echo.websocket.org

# Execute request under the 'prod' environment profile mapping base URL and auth headers
kurl --env prod /users
```

## 📚 Documentation

Detailed documentation is available in the [`docs/`](docs/) directory:

*   📖 **[Usage Guide](docs/USAGE.md)**: Detailed command-line flags, output control, and troubleshooting.
*   🏗️ **[Architecture](docs/ARCHITECTURE.md)**: Learn how `kurl` achieves fast DNS racing, smart formatting, and more.
*   🤝 **[Contributing](docs/CONTRIBUTING.md)**: Guidelines for setting up the developer environment and submitting pull requests.

## 📄 License

`kurl` is open-source software licensed under the **MIT License**.