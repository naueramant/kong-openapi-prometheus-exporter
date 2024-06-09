package swagger

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

func getPathParametersForOperationStr(operation string, pathItem *v3.PathItem) []*v3.Parameter {
	var parameters []*v3.Parameter

	// Get the parameters for the operation
	switch operation {
	case "GET":
		parameters = pathItem.Get.Parameters
	case "POST":
		parameters = pathItem.Post.Parameters
	case "PUT":
		parameters = pathItem.Put.Parameters
	case "DELETE":
		parameters = pathItem.Delete.Parameters
	case "PATCH":
		parameters = pathItem.Patch.Parameters
	case "OPTIONS":
		parameters = pathItem.Options.Parameters
	case "HEAD":
		parameters = pathItem.Head.Parameters
	default:
		return nil
	}

	// Append the parameters from the path item
	parameters = append(parameters, pathItem.Parameters...)

	return parameters
}
