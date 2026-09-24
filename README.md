# Grokipedia CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/Grokipedia-CLI)](https://goreportcard.com/report/github.com/dl-alexandre/Grokipedia-CLI)

A command-line interface for the Grokipedia API.

> **Unofficial community tool** — not affiliated with or endorsed by xAI.

> **Note:** The active implementation uses the Kong CLI framework in `internal/cli`.  
> The Cobra-based code in `cmd/` is legacy, incomplete, and no longer maintained.  
> Newer commands (`list`, `stats`, `preview`, `tts`, `random`, `edit`) only exist in the active path.

## Installation

### From Source

```bash
go install github.com/dl-alexandre/Grokipedia-CLI@latest
```

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/dl-alexandre/Grokipedia-CLI/releases).

## Quick Start

```bash
# Check the installed CLI
grokipedia version

# Search for pages
grokipedia search "python programming"

# View a specific page
grokipedia page Python_programming_language

# Get search suggestions
grokipedia typeahead "pyt"

# List edit requests
grokipedia edits

# View edit requests for a specific page
grokipedia edits-by-slug Python_programming_language

# Get API constants
grokipedia constants

# Explore randomly
grokipedia random

# Browse pages
grokipedia list --limit 10
```

## Configuration

Configuration is loaded from (in order of precedence):

1. CLI flags
2. Environment variables
3. Config file (`~/.grokipedia/config.yml`)
4. Hardcoded defaults

### Config File Example

Create `~/.grokipedia/config.yml`:

```yaml
api:
  url: "https://grokipedia.com"
  timeout: 30

cache:
  enabled: true
  ttl: 604800  # 7 days in seconds
  dir: "~/.grokipedia/cache"

output:
  format: "table"
  color: "auto"

commands:
  search:
    limit: 12
    offset: 0
    format: "table"
  edits:
    limit: 20
```

### Environment Variables

All configuration options can be set via environment variables:

- `GROKIPEDIA_API_URL` - API base URL
- `GROKIPEDIA_TIMEOUT` - Request timeout in seconds
- `GROKIPEDIA_NO_CACHE` - Set to "true" to disable caching
- `GROKIPEDIA_CACHE_DIR` - Cache directory path
- `GROKIPEDIA_CACHE_TTL` - Cache TTL in seconds
- `GROKIPEDIA_VERBOSE` - Enable verbose output
- `GROKIPEDIA_DEBUG` - Enable debug output
- `GROKIPEDIA_COLOR` - Color mode: auto, always, never

## Current API compatibility (September 2026)

The September 22, 2026 Grokipedia v0.2 announcement is a **web product
preview**; it does not announce a Grokipedia CLI v0.2 release or a stable public
API contract. The CLI remains independently versioned and currently targets
the live public endpoints, which have changed independently of that
announcement.

The current client handles the live API shapes and older compatible deployments where practical:

- Search and typeahead use the current `query` parameter (with a legacy `q` fallback for search).
- Full pages use `/api/page-preview`; `/api/page` is retained as a fallback for older deployments.
- Search and page view counts are accepted as either JSON strings or numbers.
- Typeahead returns result objects rather than a legacy `suggestions` string array.
- Edit-history results use the current `userId`, `createdAt`, and review fields.
- Article and edit submissions use the current description/evidence payload fields.

Grokipedia does not currently expose the old `/api/constants` endpoint, and the global edit feed can be unavailable; use `edits-by-slug` for per-article history. The CLI reports these cases clearly instead of treating the web preview as a CLI feature.

## Commands

### search

Search for pages in Grokipedia.

```bash
grokipedia search <query> [flags]

Flags:
  --limit int      Maximum results (1-100) (default 12)
  --offset int     Pagination offset (default 0)
  --format string  Output format: table, json, markdown (default "table")
```

### page

Retrieve a page by slug.

```bash
grokipedia page <slug> [flags]

Flags:
  --content        Show page content
  --no-links       Skip link validation
  --format string  Output format: markdown, plain, json (default "markdown")
```

### typeahead

Get search suggestions.

```bash
grokipedia typeahead <query> [flags]

Flags:
  --limit int      Maximum suggestions (1-50) (default 10)
  --format string  Output format: list, json (default "list")
```

### constants

Retrieve legacy API constants. The current Grokipedia site no longer exposes
`/api/constants`, so this command reports that endpoint as unavailable unless
it is being used with an older compatible deployment.

```bash
grokipedia constants [flags]

Flags:
  --format string  Output format: json, yaml, table (default "json")
```

### edits

List edit requests.

```bash
grokipedia edits [flags]

Flags:
  --limit int          Maximum results (1-100) (default 20)
  --status string      Filter by status (comma-separated: approved,implemented,pending)
  --exclude-user       Exclude edits by username (repeatable)
  --counts             Include count metadata (default true)
  --format string      Output format: table, json (default "table")
```

### edits-by-slug

List edit requests for a specific page.

```bash
grokipedia edits-by-slug <slug> [flags]

Flags:
  --limit int      Maximum results (1-100) (default 10)
  --offset int     Pagination offset (default 0)
  --format string  Output format: table, json (default "table")
```

### links

List internal and external links (citations) from a page.

```bash
grokipedia links <slug> [flags]

Flags:
  --internal         Show only internal page links
  --external         Show only external links (citations)
  --tree             Display links in a tree structure
  --format string    Output format: table, json, markdown, tree (default "table")
```

### suggest

Suggest a new article to be created by Grok.

```bash
grokipedia suggest <title> [flags]

Flags:
  --description string  Optional details/description for the article
  --content string      Deprecated alias for --description
  --sources string      Optional legacy sources or references
  --format string       Output format: text, json (default "text")
```

### edit

Suggest an edit to an existing article (requires xAI account sign-in via the website).

```bash
grokipedia edit <slug> [flags]

Flags:
  --summary string          Short summary of the change (required)
  --proposed-content string Proposed replacement content
  --original-content string The original text being corrected
  --section-title string    Section containing the proposed change
  --evidence string         Supporting source URL (repeatable)
  --content string          Deprecated alias for --proposed-content
  --original-text string    Deprecated alias for --original-content
  --sources string          Deprecated comma/space-separated source URLs
  --format string           Output format: text, json (default "text")
```

### list

Browse and discover pages with pagination and optional category filtering.

```bash
grokipedia list [flags]

Flags:
  --limit int       Maximum results (1-100) (default 20)
  --offset int      Pagination offset (default 0)
  --category string Filter by category
  --format string   Output format: table, json (default "table")
  --counts          Show total count and has-more info (default true)
```

### stats

Show global Grokipedia statistics.

```bash
grokipedia stats [flags]

Flags:
  --format string   Output format: table, json (default "table")
```

### preview

Lightweight page preview (useful for discovery without loading full content).

```bash
grokipedia preview <slug> [flags]

Flags:
  --format string   Output format: markdown, json, plain (default "markdown")
```

### tts

List text-to-speech sections available for reading an article aloud.

```bash
grokipedia tts <slug> [flags]

Flags:
  --format string   Output format: table, json, list (default "table")
```

### random

Show a random page (excellent for exploration and serendipity).

```bash
grokipedia random [flags]

Flags:
  --content         Fetch and display the full page content
  --format string   Output format: markdown, json, plain (default "markdown")
```

## Global Flags

These flags work with all commands:

```bash
--apiurl string       API base URL (env: GROKIPEDIA_API_URL)
--timeout int         Request timeout in seconds (env: GROKIPEDIA_TIMEOUT)
--no-cache            Disable caching (env: GROKIPEDIA_NO_CACHE)
--cache-dir string    Cache directory (env: GROKIPEDIA_CACHE_DIR)
--cache-ttl int       Cache TTL in seconds (env: GROKIPEDIA_CACHE_TTL)
-v, --verbose         Enable verbose output (env: GROKIPEDIA_VERBOSE)
--debug               Enable debug output (env: GROKIPEDIA_DEBUG)
--config-file string  Config file path (env: GROKIPEDIA_CONFIG)
--color string        Color mode: auto, always, never (env: GROKIPEDIA_COLOR)
```

## Exit Codes

- `0` - Success
- `1` - Generic error (API errors, I/O errors, config errors)
- `2` - Not found (404, empty results, unknown constant key)
- `3` - Rate limited (429 after retries)
- `4` - Invalid arguments (bad flags, unsupported format, missing required arg)

## Caching

The CLI caches API responses to improve performance. Cache files are stored in `~/.grokipedia/cache/` by default. The cache respects TTL settings and automatically invalidates expired entries.

To disable caching for a single command:
```bash
grokipedia --no-cache search "query"
```

To clear the cache:
```bash
rm -rf ~/.grokipedia/cache/
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
grokipedia-cli/
├── cmd/                    # Legacy Cobra implementation (build tag: legacy)
├── internal/
│   ├── api/               # HTTP client and models
│   ├── cache/             # File caching
│   ├── cli/               # Active Kong command implementation
│   ├── config/            # Configuration management
│   └── formatter/         # Output formatters
├── main.go                # Entry point
└── testdata/              # Test fixtures
```

## License

MIT License - see LICENSE file for details.
