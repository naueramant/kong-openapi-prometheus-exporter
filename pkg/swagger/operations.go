package swagger

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

var operations = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}

func isOperationInPathItem(pathItem *v3.PathItem, method string) bool {
	switch method {
	case "GET":
		return pathItem.Get != nil
	case "POST":
		return pathItem.Post != nil
	case "PUT":
		return pathItem.Put != nil
	case "DELETE":
		return pathItem.Delete != nil
	case "PATCH":
		return pathItem.Patch != nil
	case "OPTIONS":
		return pathItem.Options != nil
	case "HEAD":
		return pathItem.Head != nil
	default:
		return false
	}
}
