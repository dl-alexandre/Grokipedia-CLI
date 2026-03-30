package cache

import (
	"fmt"
	"testing"
)

// BenchmarkCacheKeyGeneration measures cache key generation performance
func BenchmarkCacheKeyGeneration(b *testing.B) {
	c := New("/tmp/bench-cache", 3600)
	params := map[string]interface{}{
		"q":      "python programming language tutorial",
		"limit":  100,
		"offset": 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GenerateKey("/api/full-text-search", params)
	}
}

// BenchmarkCacheKeyGenerationLargeParams measures performance with many params
func BenchmarkCacheKeyGenerationLargeParams(b *testing.B) {
	c := New("/tmp/bench-cache", 3600)
	params := make(map[string]interface{})
	for i := 0; i < 50; i++ {
		params[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GenerateKey("/api/complex-endpoint", params)
	}
}

// BenchmarkCacheSet measures cache write performance
func BenchmarkCacheSet(b *testing.B) {
	tmpDir := b.TempDir()
	c := New(tmpDir, 3600)
	data := []byte(`{"results": [{"title": "Test", "slug": "Test"}], "totalCount": 1}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		if err := c.Set(key, data); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

// BenchmarkCacheGet measures cache read performance
func BenchmarkCacheGet(b *testing.B) {
	tmpDir := b.TempDir()
	c := New(tmpDir, 3600)
	data := []byte(`{"results": [{"title": "Test", "slug": "Test"}], "totalCount": 1}`)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		if err := c.Set(key, data); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		c.Get(key)
	}
}

// BenchmarkCacheSetGet measures combined read/write performance
func BenchmarkCacheSetGet(b *testing.B) {
	tmpDir := b.TempDir()
	c := New(tmpDir, 3600)
	data := []byte(`{"test": "data"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = c.Set(key, data)
		c.Get(key)
	}
}
