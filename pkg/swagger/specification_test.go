package swagger

import (
	"context"
	"testing"

	"github.com/tj/assert"
)

func TestSpecification_MatchPath(t *testing.T) {
	ctx := context.Background()

	spec, err := LoadFile(ctx, "../../testdata/spec.yaml")
	assert.NoError(t, err)

	tests := []struct {
		method string
		path   string
		match  bool
	}{
		// Simple paths
		{"GET", "/api/v1/users", true},
		{"POST", "/api/v1/users", true},

		// All the methods
		{"GET", "/api/v1/users/1", true},
		{"PUT", "/api/v1/users/1", true},
		{"PATCH", "/api/v1/users/1", true},
		{"DELETE", "/api/v1/users/1", true},
		{"OPTIONS", "/api/v1/users/1", true},
		{"HEAD", "/api/v1/users/1", true},
		{"GET", "/api/v1/users/{userId}", false},

		// Path with multiple parameters
		{"GET", "/api/v1/users/2/posts/foo-bar", true},
		{"GET", "/api/v1/users/2/posts/foo-asd_123+324DD:33.31", true},
		{"GET", "/api/v1/users/2/posts/{postId}", false},
		{"GET", "/api/v1/users/{userId}/posts/{postId}", false},

		// All the types of parameters
		{"GET", "/api/v1/users/1/posts/foobar/comments/2/sorted/true", true},
		{"GET", "/api/v1/users/1/posts/foobar/comments/2/sorted/false", true},

		// Ubiquitous paths
		{"GET", "/api/v1/users/some-pid/info", true},
		{"GET", "/api/v1/users/john.doe@email.com/history", true},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			_, ok := spec.MatchPath(test.method, test.path)
			assert.Equal(t, test.match, ok, test.path)
		})
	}
}
