# kurl Roadmap & TODOs

This file tracks upcoming optimizations, known issues, and planned features. 
If you are looking to contribute, pick any item from this list and submit a PR!

## 🚀 Enhancements
- [ ] **HTTP/3 Support**: Investigate and add experimental QUIC/HTTP3 fallback logic in the concurrency probing system.
- [ ] **Streaming Parser**: For extremely large JSON/HTML payloads (>100MB), the current parsers buffer too much memory. Switch to stream-based rendering where tokens are flushed immediately to stdout.
- [ ] **Custom Syntax Themes**: Allow users to define a custom `.kurl.yml` config in their home directory to map their own ANSI colors to JSON/HTML tokens.
- [ ] **Websocket Probing**: Add support for `ws://` and `wss://` protocols with real-time frame formatting.

## 🐛 Known Issues (Bugs to Fix)
- [ ] **HTML Invalid Closing Tags**: The HTML pretty-printer occasionally panics or misaligns indentations on heavily malformed legacy HTML where closing tags are completely missing.
- [ ] **DNS Race Fallback**: On IPv6-only networks, the hardcoded Cloudflare `1.1.1.1` race might cause unnecessary timeouts. Implement `2606:4700:4700::1111` IPv6 fallback.

## 🧹 Technical Debt
- [ ] Improve test coverage in `color/` formatting packages.
- [ ] Move `main.go` logic to an internal `app` or `cmd` subpackage to reduce root file size.
