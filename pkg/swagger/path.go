package swagger

import "strings"

func splitPath(path string) []string {
	parts := strings.Split(path, "/")

	return parts[1:]
}
