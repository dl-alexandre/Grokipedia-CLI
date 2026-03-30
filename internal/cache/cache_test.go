package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New("/tmp/test-cache", 3600)
	if c == nil {
		t.Fatal("Expected cache instance, got nil")
	}
	if c.dir != "/tmp/test-cache" {
		t.Errorf("Expected dir '/tmp/test-cache', got %q", c.dir)
	}
	if c.ttl != 3600 {
		t.Errorf("Expected TTL 3600, got %d", c.ttl)
	}
}

func TestGenerateKey(t *testing.T) {
	c := New("/tmp/test", 3600)

	tests := []struct {
		name     string
		endpoint string
		params   map[string]interface{}
		wantLen  int // SHA256 hex is 64 chars, we take first 12
	}{
		{
			name:     "simple params",
			endpoint: "/api/search",
			params:   map[string]interface{}{"q": "test", "limit": 10},
			wantLen:  12,
		},
		{
			name:     "empty params",
			endpoint: "/api/constants",
			params:   map[string]interface{}{},
			wantLen:  12,
		},
		{
			name:     "nil params",
			endpoint: "/api/page",
			params:   nil,
			wantLen:  12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := c.GenerateKey(tt.endpoint, tt.params)
			if len(key) != tt.wantLen {
				t.Errorf("GenerateKey() length = %d, want %d", len(key), tt.wantLen)
			}
		})
	}
}

func TestGenerateKeyConsistency(t *testing.T) {
	c := New("/tmp/test", 3600)

	// Same inputs should produce same key
	params := map[string]interface{}{"q": "python", "limit": 10}
	key1 := c.GenerateKey("/api/search", params)
	key2 := c.GenerateKey("/api/search", params)

	if key1 != key2 {
		t.Errorf("GenerateKey not consistent: %q != %q", key1, key2)
	}

	// Different order should produce same key (canonicalization)
	params1 := map[string]interface{}{"a": 1, "b": 2}
	params2 := map[string]interface{}{"b": 2, "a": 1}
	key3 := c.GenerateKey("/api/test", params1)
	key4 := c.GenerateKey("/api/test", params2)

	if key3 != key4 {
		t.Errorf("GenerateKey not canonical: %q != %q", key3, key4)
	}
}

func TestSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 3600)

	key := "test-key"
	data := []byte(`{"test": "data"}`)

	// Set data
	err := c.Set(key, data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get data
	got, found := c.Get(key)
	if !found {
		t.Error("Get() returned found=false, want true")
	}
	if string(got) != string(data) {
		t.Errorf("Get() = %q, want %q", string(got), string(data))
	}
}

func TestGetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 3600)

	_, found := c.Get("non-existent-key")
	if found {
		t.Error("Get() returned found=true for non-existent key, want false")
	}
}

func TestGetExpired(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 1) // 1 second TTL

	key := "expired-key"
	data := []byte(`{"test": "data"}`)

	// Set data
	err := c.Set(key, data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Get should return not found
	_, found := c.Get(key)
	if found {
		t.Error("Get() returned found=true for expired key, want false")
	}
}

func TestGetCorruptMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 3600)

	key := "corrupt-key"
	data := []byte(`{"test": "data"}`)

	// Set data
	err := c.Set(key, data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Corrupt the metadata file
	metaPath := filepath.Join(tmpDir, key+".meta")
	err = os.WriteFile(metaPath, []byte("invalid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to corrupt metadata: %v", err)
	}

	// Get should return not found and clean up
	_, found := c.Get(key)
	if found {
		t.Error("Get() returned found=true for corrupt metadata, want false")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 3600)

	key := "delete-key"
	data := []byte(`{"test": "data"}`)

	// Set data
	err := c.Set(key, data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Delete
	err = c.Delete(key)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Get should return not found
	_, found := c.Get(key)
	if found {
		t.Error("Get() returned found=true after delete, want false")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 3600)

	// Set multiple entries
	for i := 0; i < 3; i++ {
		key := string(rune('a' + i))
		err := c.Set(key, []byte("data"))
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Clear
	err := c.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// All should be gone
	for i := 0; i < 3; i++ {
		key := string(rune('a' + i))
		_, found := c.Get(key)
		if found {
			t.Errorf("Get(%q) returned found=true after clear, want false", key)
		}
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		ttl      int
		expected bool
	}{
		{
			name:     "enabled with positive TTL",
			ttl:      3600,
			expected: true,
		},
		{
			name:     "disabled with zero TTL",
			ttl:      0,
			expected: false,
		},
		{
			name:     "disabled with negative TTL",
			ttl:      -1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New("/tmp/test", tt.ttl)
			if got := c.IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSetCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "cache", "dir")
	c := New(nestedDir, 3600)

	key := "test-key"
	data := []byte(`{"test": "data"}`)

	// Set data - should create nested directories
	err := c.Set(key, data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Set() did not create nested directories")
	}

	// Verify we can read the data
	got, found := c.Get(key)
	if !found {
		t.Error("Get() returned found=false, want true")
	}
	if string(got) != string(data) {
		t.Errorf("Get() = %q, want %q", string(got), string(data))
	}
}

func TestCacheMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, 3600)

	key := "metadata-test"
	data := []byte(`{"test": "data"}`)

	beforeSet := time.Now().Unix()
	err := c.Set(key, data)
	afterSet := time.Now().Unix()

	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Read metadata file directly
	metaPath := filepath.Join(tmpDir, key+".meta")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	var meta CacheMetadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if meta.TTL != 3600 {
		t.Errorf("Metadata TTL = %d, want 3600", meta.TTL)
	}

	if meta.CreatedAt < beforeSet || meta.CreatedAt > afterSet {
		t.Errorf("Metadata CreatedAt = %d, expected between %d and %d", meta.CreatedAt, beforeSet, afterSet)
	}
}
