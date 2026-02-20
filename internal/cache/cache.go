package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Cache handles file-based caching with TTL support
type Cache struct {
	dir string
	ttl int // seconds
}

// CacheMetadata stores cache entry metadata
type CacheMetadata struct {
	CreatedAt int64 `json:"created_at"`
	TTL       int   `json:"ttl"`
}

// New creates a new Cache instance
func New(dir string, ttl int) *Cache {
	return &Cache{
		dir: dir,
		ttl: ttl,
	}
}

// GenerateKey creates a cache key from endpoint and parameters
func (c *Cache) GenerateKey(endpoint string, params map[string]interface{}) string {
	// Sort params by key name for canonicalization
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	values := url.Values{}
	for _, k := range keys {
		if v, ok := params[k]; ok && v != nil {
			values.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Create canonical string
	canonical := endpoint + "?" + values.Encode()

	// Hash with SHA256, take first 12 chars
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])[:12]
}

// Get retrieves a cached value if it exists and is not expired
func (c *Cache) Get(key string) ([]byte, bool) {
	dataPath := filepath.Join(c.dir, key+".json")
	metaPath := filepath.Join(c.dir, key+".meta")

	// Check if data file exists
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, false
	}

	// Read and validate metadata
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		// Metadata missing, treat as expired
		os.Remove(dataPath)
		return nil, false
	}

	var meta CacheMetadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		// Corrupt metadata, delete files
		os.Remove(dataPath)
		os.Remove(metaPath)
		return nil, false
	}

	// Check if expired
	if c.ttl > 0 {
		age := time.Now().Unix() - meta.CreatedAt
		if age > int64(meta.TTL) {
			// Expired, delete files
			os.Remove(dataPath)
			os.Remove(metaPath)
			return nil, false
		}
	}

	return data, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, data []byte) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.dir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	dataPath := filepath.Join(c.dir, key+".json")
	metaPath := filepath.Join(c.dir, key+".meta")

	// Write data file
	if err := os.WriteFile(dataPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache data: %w", err)
	}

	// Write metadata file
	meta := CacheMetadata{
		CreatedAt: time.Now().Unix(),
		TTL:       c.ttl,
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		// Clean up data file on error
		os.Remove(dataPath)
		return fmt.Errorf("failed to marshal cache metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0600); err != nil {
		// Clean up data file on error
		os.Remove(dataPath)
		return fmt.Errorf("failed to write cache metadata: %w", err)
	}

	return nil
}

// Delete removes a cached entry
func (c *Cache) Delete(key string) error {
	dataPath := filepath.Join(c.dir, key+".json")
	metaPath := filepath.Join(c.dir, key+".meta")

	os.Remove(dataPath)
	os.Remove(metaPath)

	return nil
}

// Clear removes all cached entries
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) == ".json" || filepath.Ext(name) == ".meta" {
			os.Remove(filepath.Join(c.dir, name))
		}
	}

	return nil
}

// IsEnabled returns true if caching is enabled (TTL > 0)
func (c *Cache) IsEnabled() bool {
	return c.ttl > 0
}
