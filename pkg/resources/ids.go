package resources

import (
	"fmt"
	"strings"
)

func ParseCollection(path string) (string, string, string) {
	id := strings.ToLower(path)
	parts := strings.Split(id, "/providers/")
	scope := parts[0]
	resourceType := parts[1]

	return id, scope, resourceType
}

func ParseResource(path string) (string, string, string, string) {
	lower := strings.ToLower(path)
	parts := strings.Split(lower, "/providers/")
	scope := parts[0]
	parts = strings.Split(parts[1], "/")
	resourceType := strings.Join(parts[0:len(parts)-1], "/")
	name := parts[len(parts)-1]
	id := fmt.Sprintf("%s/providers/%s/%s", scope, resourceType, name)

	return id, scope, resourceType, name
}

func ParsePlaneScope(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	return "/" + strings.Join([]string{parts[0], parts[1], parts[2]}, "/")
}

func ParseNamespace(path string) string {
	_, _, resourceType, _ := ParseResource(path)
	parts := strings.Split(resourceType, "/")
	return parts[0]
}
