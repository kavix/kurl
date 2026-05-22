# 🔧 Developer & Contribution Guide

First off, thank you for considering contributing to `kurl`. It's people like you that make `kurl` such a great tool. 

## Getting Started

To get started with development, you will need a Go toolchain installed (`Go 1.22` or newer).

```bash
# Clone the repository
git clone https://github.com/kavix/kurl.git
cd kurl

# Download dependencies
go mod tidy
```

## Running Tests

`kurl` has unit test suites covering the DNS concurrent racing system, JSON tokenization, HTML collapsing, and CLI parser rules. Before submitting a PR, ensure all tests pass:

```bash
# Run all tests
make test
# Or using the go tool directly
go test -v ./...
```

## Code Formatting and Linting

We enforce standard Go formatting and linting rules to maintain a clean codebase.
If you don't have `golangci-lint` installed, you can [install it here](https://golangci-lint.run/usage/install/).

Before committing, run:
```bash
# Format your code
make fmt

# Run the linter
make lint
```

## Compiling Production Binaries

To build a highly optimized binary locally with stripped debug symbols and optimized size:

```bash
make build
```
The binary will be generated as `kurl` in the root directory.

To install it directly to your `~/.local/bin`:
```bash
make install
```

## Snapshot Compilation (Multi-Platform)

If you want to test cross-compilation on your machine for macOS, Linux, and Windows (using GoReleaser):

```bash
make release-local
```

## Pull Request Process

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Code of Conduct
Please be respectful and considerate of others when contributing. We welcome contributions from everyone.
