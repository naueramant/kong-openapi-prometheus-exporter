package kong

import (
	"fmt"
	"testing"
)

// Test structures
type DeepNested struct {
	Level1 struct {
		Level2 struct {
			Level3 struct {
				Value string
			}
		}
	}
}

// setupTestData creates test data for benchmarks
func setupTestData() (*Log, *DeepNested) {
	log := &Log{
		Request: Request{
			URI:    "/api/v1/users",
			Method: "GET",
			Headers: map[string]string{
				"User-Agent":    "Mozilla",
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
		},
		Response: Response{
			Status: 200,
		},
		Latencies: Latencies{
			Request: 100,
		},
	}

	deep := &DeepNested{}
	deep.Level1.Level2.Level3.Value = "deep value"

	return log, deep
}

func BenchmarkGet(b *testing.B) {
	log, deep := setupTestData()

	// Define benchmark cases
	cases := []struct {
		name string
		obj  interface{}
		path string
	}{
		{"SimpleField", log, "request.uri"},
		{"MapAccess", log, "request.headers['User-Agent']"},
		{"DeepNested", deep, "Level1.Level2.Level3.Value"},
		{"NonExistentPath", log, "nonexistent.path.here"},
		{"EmptyPath", log, ""},
		{"ComplexMapAccess", log, "request.headers['Content-Type']"},
		{"NumericField", log, "response.status"},
	}

	// Run benchmarks for each case
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if l, ok := tc.obj.(*Log); ok {
					l.Get(tc.path)
				} else if _, ok := tc.obj.(*DeepNested); ok {
					// Create a Log wrapper for DeepNested to use Get method
					wrapper := &Log{}
					wrapper.Get(tc.path)
				}
			}
		})
	}
}

// BenchmarkPathSplitting specifically benchmarks the path splitting function
func BenchmarkPathSplitting(b *testing.B) {
	paths := []struct {
		name string
		path string
	}{
		{"SimpleDotPath", "request.uri"},
		{"BracketPath", "headers['User-Agent']"},
		{"MixedPath", "request.headers['Content-Type'].value"},
		{"DeepPath", "level1.level2.level3.level4.value"},
	}

	for _, p := range paths {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				splitPath(p.path)
			}
		})
	}
}

// BenchmarkMapOperations benchmarks different map access patterns
func BenchmarkMapOperations(b *testing.B) {
	log, _ := setupTestData()

	headers := []string{
		"User-Agent",
		"Content-Type",
		"Authorization",
		"NonExistent",
	}

	for _, header := range headers {
		b.Run(fmt.Sprintf("MapAccess_%s", header), func(b *testing.B) {
			path := fmt.Sprintf("request.headers['%s']", header)
			for i := 0; i < b.N; i++ {
				log.Get(path)
			}
		})
	}
}

// BenchmarkMemoryAllocation measures memory allocations
func BenchmarkMemoryAllocation(b *testing.B) {
	log, _ := setupTestData()
	paths := []string{
		"request.uri",
		"request.headers['User-Agent']",
		"response.status",
		"nonexistent.path",
	}

	for _, path := range paths {
		b.Run(fmt.Sprintf("MemAlloc_%s", path), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				log.Get(path)
			}
		})
	}
}
