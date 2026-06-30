# kurl

<img width="2816" height="1536" alt="Gemini_Generated_Image_p38qe0p38qe0p38q" src="https://github.com/user-attachments/assets/79afaf1c-ddde-44bf-b545-c19a1adda095" />

> A modern, concurrent HTTP client for the terminal — built in Go. Zero-config scheme probing, DNS racing, smart formatting, request replays, WebSockets, and environment profiles. All in one static binary.

```
# Fetch any API — kurl auto-detects scheme, races DNS, and pretty-prints JSON
kurl api.github.com/users/kavix

# POST with headers and body
kurl POST https://api.example.com/v1/users   -H "Content-Type: application/json"   -d '{"name": "Alice", "role": "admin"}'

# Save a request profile and replay it
kurl save github-api GET https://api.github.com/users/kavix -H "Accept: application/json"
kurl run github-api -v

# Interactive WebSocket session
kurl ws://echo.websocket.org

# Switch environments on the fly
kurl --env prod /users
```

## What is kurl?

kurl is a command-line HTTP client that eliminates the friction of debugging APIs in the terminal. Unlike curl, which dumps raw, unformatted output, kurl renders responses with strict indentation, syntax highlighting, and structured headers — while adding modern features like concurrent DNS racing, automatic scheme probing, request replay profiles, and WebSocket support.

It is designed for developers who want curl's universality without its steep learning curve, and HTTPie's beauty without its Python dependency.

## Why kurl?

| Feature | curl | HTTPie | kurl |
|---------|------|--------|------|
| Single static binary | ✅ | ❌ (Python) | ✅ |
| Auto scheme probing (http/https) | ❌ | ❌ | ✅ |
| Concurrent DNS racing | ❌ | ❌ | ✅ |
| Token-by-token JSON formatting | ❌ | ✅ | ✅ |
| Smart HTML pretty-printing | ❌ | ❌ | ✅ |
| Anti-bot header injection | ❌ | ❌ | ✅ |
| Request save/replay profiles | ❌ | ❌ | ✅ |
| Interactive WebSockets | ❌ | ❌ | ✅ |
| Environment profiles (dev/staging/prod) | ❌ | ❌ | ✅ |
| Zero runtime dependencies | ✅ | ❌ | ✅ |

## Technical Highlights

- **Concurrent DNS Racing**: Queries system DNS and Cloudflare's `1.1.1.1` in parallel to eliminate VPN-induced DNS hang latencies.
- **Zero-Config Scheme Probing**: Pass a bare domain (`kurl example.com`) and kurl probes `https://` and `http://` concurrently, returning the fastest successful response.
- **Token-by-Token JSON Formatter**: Parses JSON on the fly with strict indentation and harmonized syntax highlighting — no external `python -m json.tool` or `jq` required.
- **HTML5 DOM Pretty-Printer**: Collapses inline element nodes to avoid line bloat while preserving structural clarity.
- **Anti-Bot Bypass**: Automatically injects standard modern browser headers to prevent WAF and anti-bot layers from rejecting CLI requests.
- **Request Replay System**: Save request configurations locally (like a terminal-native Postman) and replay them with optional parameter overrides.
- **Interactive WebSocket Client**: Full-duplex text frame communication with real-time colorized message formatting.
- **Environment Profiles**: Switch base URLs, auth headers, and defaults dynamically between dev, staging, and production via local configuration files.

## Installation

### macOS / Linux (Homebrew)

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

> Ensure `~/.local/bin` is in your `PATH`.

## Quick Start

```bash
# Fetch and format a JSON API
kurl https://api.genderize.io/?name=luc

# Fetch a webpage with automatic scheme probing and smart HTML rendering
kurl news.lk

# POST JSON payload with custom headers
kurl POST https://api.example.com/v1/users   -H "Content-Type: application/json"   -H "Authorization: Bearer my-secret-token"   -d '{"name": "Alice", "role": "admin"}'

# Save a request configuration locally
kurl save github-api GET https://api.github.com/users/kavix -H "Accept: application/json"

# Replay the request with optional parameter overrides
kurl run github-api -v

# Open an interactive WebSocket session with colorized frames
kurl ws://echo.websocket.org

# Execute request under the 'prod' environment profile
kurl --env prod /users
```

## Documentation

- [Usage Guide](docs/usage.md) — Command-line flags, output control, and troubleshooting
- [Architecture](docs/architecture.md) — DNS racing, smart formatting, and design decisions
- [Contributing](CONTRIBUTING.md) — Developer setup and pull request guidelines

## License

MIT License — see [LICENSE](LICENSE) for details.
