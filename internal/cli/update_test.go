package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAutoUpdateCacheDirUsesCacheDir(t *testing.T) {
	t.Setenv("CACHE_DIR", "/tmp/grokipedia-test-cache")

	got := autoUpdateCacheDir("grokipedia")
	want := filepath.Join("/tmp/grokipedia-test-cache", "grokipedia")
	if got != want {
		t.Fatalf("autoUpdateCacheDir() = %q, want %q", got, want)
	}
}

func TestHasFreshUpdateCache(t *testing.T) {
	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "update_check.json")

	writeCache := func(createdAt time.Time) {
		data, err := json.Marshal(map[string]time.Time{"created_at": createdAt})
		if err != nil {
			t.Fatalf("failed to marshal cache: %v", err)
		}
		if err := os.WriteFile(cachePath, data, 0o644); err != nil {
			t.Fatalf("failed to write cache: %v", err)
		}
	}

	writeCache(time.Now())
	if !hasFreshUpdateCache(cacheDir, time.Hour) {
		t.Fatal("expected a newly written update cache to be fresh")
	}

	writeCache(time.Now().Add(-2 * time.Hour))
	if hasFreshUpdateCache(cacheDir, time.Hour) {
		t.Fatal("expected an expired update cache to be rejected")
	}
}
