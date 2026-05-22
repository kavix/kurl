# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Created `docs/` directory for detailed documentation architecture (`USAGE.md`, `ARCHITECTURE.md`, `CONTRIBUTING.md`).
- Added issue templates for Bug Reports and Feature Requests.
- Added Pull Request template to standardize external contributions.
- Added GitHub actions and workflows for release automation.

### Changed
- Refactored `README.md` to be a cleaner, professional landing page.

## [1.0.0] - 2026-05-22

### Added
- Initial release of `kurl`.
- Concurrency-powered probing for `http://` and `https://`.
- Concurrent DNS Racing to bypass bottlenecks.
- Smart Token-by-Token JSON Formatter with strict indentation and syntax highlighting.
- Smart HTML Pretty-Printer utilizing an HTML5 DOM parser with inline element collapsing.
- Auto-injection of standard browser headers (`User-Agent`, `Accept`) to bypass anti-bot protections.
- Smart TTY Output Switching to seamlessly strip ANSI colors on redirect.
- Verbose Request Chains (`-v` flag) for tracking full redirect hops and absolute headers.
- Homebrew Tap integration for cross-platform macOS/Linux installation.
