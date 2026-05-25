# Architecture & Technical Design

This document details the core architectural features that make `kurl` fast, robust, and beautiful.

## Concurrency-Powered Probing

When you pass a raw domain (e.g. `google.com`) without a scheme to `kurl`, it automatically initiates parallel HTTP probes. `kurl` will query both `https://` and `http://` concurrently. Whichever protocol responds successfully first is automatically resolved and served to the user. This guarantees the fastest possible resolution and gracefully handles environments that only support HTTP.

## Concurrent DNS Racing Resolver

Standard DNS resolution relies on the host's system configuration, which can sometimes be bottlenecked by VPNs, ISP restrictions, or slow local resolvers.
`kurl` bypasses these bottlenecks entirely using a **Multi-threaded DNS racing** approach. 

For every hostname lookup, `kurl` simultaneously dispatches queries to:
1. The default system DNS resolver.
2. A fast public resolver (like Cloudflare's `1.1.1.1`).

The query that completes first wins the race, and its IP result is used to dial the connection. This design completely eliminates DNS hang latencies and provides incredible resilience.

## Token-by-Token JSON Formatter

Traditional JSON formatting often loads the entire JSON document into a map/struct and re-encodes it. This can lose ordering and be relatively slow for huge payloads.
Instead, `kurl` uses a strict **token-by-token parsing engine**. It parses JSON response bodies on the fly, rendering them directly to the terminal buffer with strict indentation and harmonized syntax-highlighting terminal colors.

Color Map:
*   **Braces/Brackets**: Gray/Dim
*   **Keys**: Cyan
*   **Strings**: Green
*   **Numbers & Booleans**: Yellow
*   **Nulls**: Bold Red

## Smart HTML Pretty-Printer

For raw HTML bodies, `kurl` uses a full HTML5 compliant DOM parser to process the structure. 

Key features include:
*   **2-space Indentations**: Clean vertical structure.
*   **Inline Element Collapsing**: Instead of placing every tag on a new line (which causes extreme vertical bloat), it intelligently collapses inline element nodes (`<b>`, `<i>`, `<a>`, `<span>`, `<strong>`, etc.) onto single lines with their text content.
*   **Syntax Highlighting**: Colorizes tags, attributes, values, and comments.

Color Map:
*   **Tags & Bracket Closures**: Cyan & Dim Gray
*   **Attributes Keys**: Yellow
*   **Attribute Values**: Green
*   **DOCTYPE Definitions**: Bold Magenta
*   **HTML Comments**: Dim Gray / Italic

## Smart TTY Output Switching

`kurl` respects the execution environment out-of-the-box. It automatically detects whether `stdout` is a terminal or if it is being redirected to a file, pipe, or script.
If the output is redirected, `kurl` seamlessly and silently strips all ANSI escape codes (colors/bolding), ensuring that plain text output is perfectly clean. It also natively supports the standard `NO_COLOR` environment variable.

## CDN & Anti-Bot Bypass
Many modern APIs and web servers sit behind CDN firewalls (Cloudflare, Akamai, etc.) that reject requests lacking standard headers. `kurl` preemptively injects standard modern browser headers (`User-Agent` and `Accept`) to bypass these anti-bot blocking layers, resulting in higher success rates for CLI-based requests out of the box.

## Request Replays & Serialization
`kurl` stores profiles under `~/.kurl/requests/<name>.json` using a custom `savedRequest` schema.
*   **Merger Parser**: Replaying configurations merges base configurations with CLI options dynamically. When parsing override arguments, the parser is seeded with the loaded configuration, allowing CLI values to replace or append to the loaded parameters.
*   **Validation**: Sanitizes file path names to prevent directory traversal attacks (e.g. `save ../../bad`).

## Interactive WebSocket Duplex Engine
When routing connections to `ws://` or `wss://`, `kurl` bypasses standard HTTP fetching and launches an asynchronous duplex network loop:
*   **Asynchronous Reader**: Spawns a background goroutine to read frames from the WebSocket stream, checking if message payloads are valid JSON to format them dynamically via the token-by-token highlighter.
*   **Sender loop**: Uses the main thread to scan standard input and send text messages directly to the socket connection.

## Environment Variable Profile Mapping
Enabling `--env <profile>` loads configuration mappings from `~/.kurl/environments.json`.
*   **Case-insensitive Header Merging**: Profile headers are merged with CLI overrides. If a header key passed on the CLI overrides a profile header, the profile header is replaced in-place, keeping ordering and avoiding duplicates.
*   **Safe URL Joiner**: Safely concatenates environment base URLs with target relative paths, stripping duplicate slashes cleanly.

