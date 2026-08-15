# Architecture & Technical Design

This document details the software architecture, runtime flow, and concurrent sub-systems of `kurl`.

---

## 1. High-Level Architecture Overview

The diagram below illustrates the end-to-end processing pipeline of `kurl`, from CLI command parsing and profile resolution down to concurrent DNS racing, network transports, protocol sub-engines (HTTP, WebSockets, SSE, GraphQL), and smart output formatting.

```mermaid
flowchart TD
    subgraph CLI ["1. CLI & Environment Layer"]
        Input["Terminal Input<br/><code>kurl [METHOD] [URL] [flags]</code>"] --> CLI_Parser["CLI Argument & Flag Parser<br/><code>main.go</code>"]
        CLI_Parser --> Env_Loader["Environment Profile Engine<br/><code>env.go (~/.kurl/environments.json)</code>"]
        CLI_Parser --> Request_Store["Request Replay Manager<br/><code>(~/.kurl/requests/*.json)</code>"]
    end

    subgraph Router ["2. Sub-Protocol Router"]
        Env_Loader --> Protocol_Switch{"Protocol / Command Dispatcher"}
        Protocol_Switch -- "HTTP / HTTPS" --> HTTP_Engine["HTTP Fetch Engine<br/><code>client/client.go</code>"]
        Protocol_Switch -- "ws:// / wss://" --> WS_Engine["WebSocket Duplex Loop<br/><code>websocket.go</code>"]
        Protocol_Switch -- "sse" --> SSE_Engine["Server-Sent Events Streamer<br/><code>internal/sse</code>"]
        Protocol_Switch -- "graphql" --> GQL_Engine["GraphQL Execution Engine<br/><code>internal/graphql</code>"]
    end

    subgraph Network ["3. Concurrent Networking Core"]
        HTTP_Engine --> DNS_Race["Triple-DNS Racing Resolver<br/>(1.1.1.1, [2606:4700:4700::1111], System)"]
        HTTP_Engine --> Scheme_Race["Scheme Probing Engine<br/>(Parallel https:// vs http://)"]
        DNS_Race --> Tuned_Transport["HTTP/2 & HTTP/1.1 Tuned Transport<br/>(Keep-Alive, TLS Handshake)"]
        Scheme_Race --> Tuned_Transport
    end

    subgraph Pipeline ["4. Rendering & Transformation Pipeline"]
        Tuned_Transport --> Filter_Engine["JSON Transformation Engine<br/><code>internal/filter (--filter, --filter-keys)</code>"]
        Filter_Engine --> Formatter{"Content-Type & TTY Auto-Detection"}
        Formatter -- "application/json" --> JSON_Printer["Token-by-Token JSON Formatter<br/><code>printer/json.go</code>"]
        Formatter -- "text/html" --> HTML_Printer["HTML5 DOM Pretty-Printer<br/><code>printer/html.go</code>"]
        Formatter -- "Raw / Redirected" --> Raw_Printer["Plain Text / Binary Stream"]
    end

    subgraph Output ["5. Terminal & File Output"]
        JSON_Printer --> Terminal["Colorized Stdout"]
        HTML_Printer --> Terminal
        Raw_Printer --> File_Save["File Output (-o / --output)"]
    end
```

---

## 2. Multi-Threaded Concurrent Resolver Architecture

`kurl` eliminates VPN and local DNS hang latencies by dispatching 3 concurrent resolution requests for every domain name. Whichever resolver returns a valid IP first cancels the remaining goroutine contexts and immediately proceeds to socket connection dialing.

```mermaid
sequenceDiagram
    autonumber
    participant Client as kurl Client
    participant Quad4 as Cloudflare IPv4 (1.1.1.1)
    participant Quad6 as Cloudflare IPv6 ([2606:4700:4700::1111])
    participant System as System Local Resolver
    participant Dial as DialContext

    Client->>Quad4: LookupIP("api.example.com", 1.1.1.1:53)
    Client->>Quad6: LookupIP("api.example.com", [2606:4700:4700::1111]:53)
    Client->>System: LookupIP("api.example.com", DefaultResolver)
    
    note over Client,System: Concurrent Race (First Successful Result Wins)
    Quad4-->>Client: Return IP [104.16.12.3] (Fastest - 4ms)
    note right of Client: Cancel Quad6 & System Contexts
    Client->>Dial: Establish TCP + TLS Connection to 104.16.12.3
```

---

## 3. Scheme Probing Engine Workflow

When passed a bare domain (`kurl example.com`), `kurl` executes concurrent HTTP and HTTPS probes to eliminate scheme guessing latencies:

```mermaid
stateDiagram-v2
    [*] --> ParseTarget
    ParseTarget --> ExplicitScheme: Has http:// or https://
    ParseTarget --> ProbeConcurrent: Bare Domain (e.g. api.org)

    state ProbeConcurrent {
        [*] --> SpawnHTTPS: Goroutine 1 (https://)
        [*] --> SpawnHTTP: Goroutine 2 (http://)
        SpawnHTTPS --> RaceBarrier
        SpawnHTTP --> RaceBarrier
    }

    RaceBarrier --> WinnerSelected: First Successful 2xx/3xx/4xx Response
    WinnerSelected --> CancelLoser: Cancel Loser Context
    ExplicitScheme --> ExecuteSingle: Direct Connection
    CancelLoser --> StreamResponse: Read & Pretty-Print Payload
    ExecuteSingle --> StreamResponse
    StreamResponse --> [*]
```

---

## 4. Sub-System Details & Industrial Design Standards

### A. Zero-Allocation Token-by-Token JSON Formatter
Traditional JSON formatters unmarshal entire documents into memory maps or struct trees before pretty-printing, resulting in high memory churn and loss of original key ordering. `kurl` utilizes an **on-the-fly streaming token parser**:
* **Allocation Caching**: Uses a pre-allocated depth indentation buffer array (`[32][]byte`) to eliminate string formatting allocations (`fmt.Sprintf`) per JSON key/val token.
* **Color Harmony**: Highlighting tokens (`Cyan` keys, `Green` strings, `Yellow` numbers/booleans, `Bold Red` nulls) streamed directly via `io.WriteString`.

### B. Smart HTML5 DOM Pretty-Printer
* **Inline Node Collapsing**: Detects inline tags (`<b>`, `<i>`, `<a>`, `<span>`, etc.) and collapses them onto single lines to prevent vertical output bloat.
* **Malformed HTML Guarding**: Maintains AST stack depth boundaries to safely format unclosed or legacy HTML tags without panic underflows.

### C. JSON Path Transformation Engine (`internal/filter`)
Provides jq-like filtering directly inside `kurl`:
* **Path Slicing (`--filter`)**: Evaluates property paths (`.users[0].name`).
* **CSV Projection (`--filter-keys`)**: Filters keys from objects or arrays (`name, email, role`).
* **Array Flattening (`--filter-flatten`)**: Unnests multi-dimensional array payloads before formatting.

### D. Server-Sent Events (SSE) Streaming Engine (`internal/sse`)
* **W3C EventSource Standard**: Parses multiline `data:` buffers, `event:` classifications, `id:` tracking tags, and `retry:` interval durations.
* **Real-time Colorized Terminal Output**: Live streams events with high-resolution timestamps (`15:04:05`) and optional event-type filtering (`--sse-filter`).

### E. GraphQL Native Integration (`internal/graphql`)
* **Payload Serialization**: Wraps queries and variable JSON payloads (`--variables`) into standard POST HTTP requests.
* **Introspection & Code Generation**: Auto-generates template queries via `--generate-query <Type>` and executes full schema introspection (`--introspect`).

---

## 5. Security & TTY Environment Control

* **Directory Traversal Protection**: Replay request profile names (`kurl save <name>`) are sanitized with strict alphanumeric, hyphen, and underscore validation rules (`isValidRequestName`).
* **Smart TTY Detection**: Uses `os.ModeCharDevice` check on `stdout`. When redirected to files or pipes, ANSI color codes are stripped silently to ensure clean text output. Natively supports the `NO_COLOR` standard.

