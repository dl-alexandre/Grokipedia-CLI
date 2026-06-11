# Grokipedia CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/grokipedia/cli)](https://goreportcard.com/report/github.com/grokipedia/cli)

A command-line interface for the Grokipedia API.

> **Unofficial community tool** — not affiliated with or endorsed by xAI.

> **Note:** The active implementation uses the Kong CLI framework in `internal/cli`.  
> The Cobra-based code in `cmd/` is legacy, incomplete, and no longer maintained.  
> Newer commands (`list`, `stats`, `preview`, `tts`, `random`, `edit`) only exist in the active path.

## Installation

### From Source

```bash
go install github.com/grokipedia/cli@latest
```

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/grokipedia/cli/releases).

## Quick Start

```bash
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
  --limit int      Maximum suggestions (1-50) (default 5)
  --format string  Output format: list, json (default "list")
```

### constants

Retrieve API constants.

```bash
grokipedia constants [flags]

Flags:
  --key string     Filter to a single constant key
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
  --content string   Optional content or details for the article
  --sources string   Optional sources or references
  --format string    Output format: text, json (default "text")
```

### edit

Suggest an edit to an existing article (requires xAI account sign-in via the website).

```bash
grokipedia edit <slug> [flags]

Flags:
  --summary string       Short summary of the change (required)
  --content string       Proposed replacement content
  --sources string       Supporting sources
  --original-text string The original text being corrected
  --format string        Output format: text, json (default "text")
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
--api-url string      API base URL (env: GROKIPEDIA_API_URL)
--timeout int         Request timeout in seconds (env: GROKIPEDIA_TIMEOUT)
--no-cache            Disable caching (env: GROKIPEDIA_NO_CACHE)
--cache-dir string    Cache directory (env: GROKIPEDIA_CACHE_DIR)
--cache-ttl int       Cache TTL in seconds (env: GROKIPEDIA_CACHE_TTL)
-v, --verbose         Enable verbose output (env: GROKIPEDIA_VERBOSE)
--debug               Enable debug output (env: GROKIPEDIA_DEBUG)
--config string       Config file path (env: GROKIPEDIA_CONFIG)
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
├── cmd/                    # Cobra commands
├── internal/
│   ├── api/               # HTTP client and models
│   ├── cache/             # File caching
│   ├── config/            # Configuration management
│   └── formatter/         # Output formatters
├── main.go                # Entry point
└── testdata/              # Test fixtures
```

## License

MIT License - see LICENSE file for details.
