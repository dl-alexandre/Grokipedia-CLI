package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dl-alexandre/cli-tools/update"
	"github.com/dl-alexandre/cli-tools/version"
)

const updateCacheTTL = 24 * time.Hour

// UpdateCheckCmd wraps cli-tools update functionality
type UpdateCheckCmd struct {
	Force bool `help:"Force check, bypassing cache" flag:"force"`
}

// Run executes the update check
func (c *UpdateCheckCmd) Run(globals *Globals) error {
	checker := update.New(update.Config{
		CurrentVersion: version.Version,
		BinaryName:     version.BinaryName,
		GitHubRepo:     "dl-alexandre/Grokipedia-CLI",
		InstallCommand: "brew upgrade grokipedia",
	})

	info, err := checker.Check(c.Force)
	if err != nil {
		return err
	}

	return update.DisplayUpdate(info, version.BinaryName, "table")
}

// AutoUpdateCheck performs a background update check for use at startup.
// Notifications go to stderr so machine-readable command output remains valid.
func AutoUpdateCheck() {
	binaryName := version.BinaryName
	if binaryName == "" {
		binaryName = "grokipedia"
	}

	cacheDir := autoUpdateCacheDir(binaryName)
	if hasFreshUpdateCache(cacheDir, updateCacheTTL) {
		return
	}

	checker := update.New(update.Config{
		CurrentVersion: version.Version,
		BinaryName:     binaryName,
		GitHubRepo:     "dl-alexandre/Grokipedia-CLI",
		CacheDir:       cacheDir,
		CacheTTL:       updateCacheTTL,
		InstallCommand: "brew upgrade grokipedia",
	})

	go func() {
		info, err := checker.Check(false)
		if err != nil || !info.UpdateAvailable {
			return
		}

		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "📦 A new version is available: %s (current: %s)\n", info.LatestVersion, info.CurrentVersion)
		fmt.Fprintf(os.Stderr, "   Run '%s check-updates' for details or upgrade with: brew upgrade grokipedia\n", binaryName)
		fmt.Fprintln(os.Stderr)
	}()
}

func autoUpdateCacheDir(binaryName string) string {
	if dir := os.Getenv("CACHE_DIR"); dir != "" {
		return filepath.Join(dir, binaryName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", binaryName)
	}
	return filepath.Join(homeDir, ".cache", binaryName)
}

func hasFreshUpdateCache(cacheDir string, ttl time.Duration) bool {
	if cacheDir == "" || ttl <= 0 {
		return false
	}

	root, err := os.OpenRoot(cacheDir)
	if err != nil {
		return false
	}
	defer root.Close()

	data, err := root.ReadFile("update_check.json")
	if err != nil {
		return false
	}

	var entry struct {
		CreatedAt time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return false
	}

	return time.Since(entry.CreatedAt) <= ttl
}

// UpdateInfo is re-exported from cli-tools for backward compatibility
type UpdateInfo = update.Info
