package swagger

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

func getPathParametersForOperationStr(operation string, pathItem *v3.PathItem) []*v3.Parameter {
	// TODO: add support for params on the path item itself

	switch operation {
	case "GET":
		return pathItem.Get.Parameters
	case "POST":
		return pathItem.Post.Parameters
	case "PUT":
		return pathItem.Put.Parameters
	case "DELETE":
		return pathItem.Delete.Parameters
	case "PATCH":
		return pathItem.Patch.Parameters
	case "OPTIONS":
		return pathItem.Options.Parameters
	case "HEAD":
		return pathItem.Head.Parameters
	case "TRACE":
		return pathItem.Trace.Parameters
	default:
		return nil
	}
}
