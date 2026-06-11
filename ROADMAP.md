# Grokipedia CLI Roadmap

This document outlines the current state, completed work, known gaps, and future direction of the Grokipedia CLI.

## Current Status (as of May 2026)

The CLI is in a **mature and usable state** for most public Grokipedia functionality.

- **Active implementation**: Located in `internal/cli` (uses the Kong CLI framework).
- **Legacy implementation**: Located in `cmd/` (Cobra-based). This is deprecated and no longer maintained.
- All major public API endpoints are supported.

### Supported Commands

| Command          | Status     | Notes |
|------------------|------------|-------|
| `search`         | Stable     | Full-text search with pagination |
| `page`           | Stable     | Full page retrieval with caching |
| `list`           | Stable     | Browse pages with category filtering + caching |
| `preview`        | Stable     | Lightweight page preview + caching |
| `random`         | Stable     | Random page discovery (supports `--category`) |
| `stats`          | Stable     | Global statistics + caching |
| `tts`            | Stable     | List TTS sections for an article + caching |
| `typeahead`      | Stable     | Search suggestions |
| `edits`          | Stable     | List edit requests |
| `edits-by-slug`  | Stable     | Edit requests for a specific page |
| `suggest`        | Stable     | Suggest new articles |
| `edit`           | Stable     | Suggest edits to existing articles (requires auth on website) |
| `constants`      | Stable     | API constants and enums |
| `links`          | Stable     | Extract links from a page |

## Completed Work

- Initial feature implementation (search, page, edits, etc.)
- Addition of discovery commands (`list`, `stats`, `preview`, `random`, `tts`)
- Contribution commands (`suggest`, `edit`)
- Caching support for expensive endpoints (`list`, `preview`, `stats`, `tts`)
- Output function extraction for maintainability
- Deprecation of the old Cobra-based CLI (`cmd/`)
- Comprehensive API client and model testing
- CLI output testing
- Legal disclaimers and clear unofficial status

## Known Gaps & Limitations

### API Coverage
- No support for user authentication / private features (activity, personal suggestions)
- No article history or version viewing
- No bulk operations or exports

### TTS
- `tts` command only lists sections. Actual audio playback is **out of scope** for now (the API currently only returns section metadata).

### Testing
- Full end-to-end command testing is limited (Run methods are harder to mock cleanly).
- No performance or load testing.

### Legacy Code
- The Cobra implementation in `cmd/` is deprecated and incomplete.
- As of May 2026, it is isolated behind `//go:build legacy` (does not compile by default).
- **Removal Plan (Proposed)**:
  1. May–June 2026: Add build tags + deprecation warnings (done)
  2. July–August 2026: Move valuable tests out of `cmd/`, strengthen warnings in code and docs
  3. September 2026: Announce removal timeline
  4. Q4 2026 / v1.0: Delete the `cmd/` directory entirely
- New contributors must only modify code in `internal/cli/`.
- If you need to run legacy tests: `go test -tags=legacy ./cmd/...`

## Future Ideas (Not Prioritized)

- Better testability for command `Run()` methods (dependency injection)
- `grokipedia info` or `doctor` command
- Improved `random` command (true randomness across more pages, reproducibility)
- Category browsing / tree view
- Shell completion improvements
- Configuration profiles
- Support for authenticated write operations (if/when the API supports it cleanly)

## Non-Goals (Out of Scope)

- Implementing actual TTS audio playback
- Replacing or competing with the official web experience
- Heavy authentication flows (login, token management)

## Versioning & Stability

- The project follows semantic versioning.
- Breaking changes will be communicated clearly.
- The legacy `cmd/` package may be removed in a future major release.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on adding new commands, testing expectations, and the development process.

---

*Last updated: May 2026*