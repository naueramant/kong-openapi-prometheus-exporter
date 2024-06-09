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
		// simple paths
		{"GET", "/api/v1/users", true},
		{"POST", "/api/v1/users", true},

		// all supported http methods
		{"GET", "/api/v1/users/1", true},
		{"PUT", "/api/v1/users/1", true},
		{"PATCH", "/api/v1/users/1", true},
		{"DELETE", "/api/v1/users/1", true},
		{"OPTIONS", "/api/v1/users/1", true},
		{"HEAD", "/api/v1/users/1", true},
		{"GET", "/api/v1/users/{userId}", false},

		// path with multiple parameters
		{"GET", "/api/v1/users/2/posts/foo-bar", true},
		{"GET", "/api/v1/users/2/posts/foo-asd_123+324DD:33.31", true},
		{"GET", "/api/v1/users/{userId}/posts/{postId}", false},

		// parameters with all data types
		{"GET", "/api/v1/users/1/posts/foobar/comments/2/sorted/true", true},
		{"GET", "/api/v1/users/1/posts/foobar/comments/2/sorted/false", true},
		{"GET", "/api/v1/users/1/posts/foobar/comments/2/sorted/1", false},
		{"GET", "/api/v1/users/1/posts/foobar/comments/2/sorted/0", false},

		// paths with multiple parameters matching the same regex
		{"GET", "/api/v1/workers/john-doe/info", true},
		{"GET", "/api/v1/workers/john.doe@email.com/history", true},

		// weird paths
		{"GET", "/api/v1/weird/foo/bar/buzz", true},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			_, ok := spec.MatchPath(test.method, test.path)
			assert.Equal(t, test.match, ok, test.path)
		})
	}
}
