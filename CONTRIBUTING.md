# Contributing to Grokipedia CLI

Thank you for your interest in contributing! This document explains how to contribute effectively.

## Project Structure

- `main.go` — Entry point. Uses the Kong-based CLI.
- `internal/cli/` — **Active implementation** (all new development happens here). Uses the Kong framework.
- `internal/api/` — HTTP client and response models for the Grokipedia API.
- `internal/cache/` — File-based response caching.
- `cmd/` — **Legacy Cobra implementation** (deprecated, do not add new features here).
- `testdata/fixtures/` — Mock API responses for testing.

## Adding a New Command

1. Create a new command struct in `internal/cli/cli.go` (following the pattern of existing commands like `ListCmd` or `StatsCmd`).
2. Implement the `Run(globals *Globals) error` method.
3. If the command makes API calls, add the corresponding method to `internal/api/client.go`.
4. Add response/request models in `internal/api/models.go` if needed.
5. Add caching where appropriate (highly recommended for read-heavy commands).
6. Extract an `output*Results` function for formatting (improves testability).
7. Add tests:
   - API client tests in `internal/api/client_test.go`
   - Output formatting tests in `internal/cli/`
8. Update `ROADMAP.md` if the command represents a significant new capability.
9. Update `README.md` with usage examples.

## Testing Guidelines

- All new API client methods must have tests using `httptest`.
- Prefer testing pure output functions (`outputListResults`, etc.) over full command execution when possible.
- For full `Run` method testing, use a test `Globals` with a mock client (see existing patterns).
- Run `go test ./...` before submitting a PR.

## Code Style

- Follow existing patterns for caching, error handling, and output formatting.
- Keep command logic focused — move formatting into dedicated `output*` functions.
- Use clear, user-friendly error messages.

## Deprecation Policy

The old Cobra-based CLI in `cmd/` is deprecated. Do not add new commands or features there. It exists only for backward compatibility during the transition period and may be removed in a future major release.

## Pull Request Process

1. Fork the repository and create a feature branch.
2. Make your changes following the guidelines above.
3. Ensure all tests pass (`go test ./...`).
4. Update documentation (README, ROADMAP if needed).
5. Open a Pull Request with a clear description of the change.

## Questions?

Feel free to open an issue for discussion before starting larger changes.

---

Thank you for helping improve the Grokipedia CLI!