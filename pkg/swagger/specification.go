package swagger

import (
	"context"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type TreeNode struct {
	Children map[string]*TreeNode

	IsParameter bool
	CanBeLeaf   bool

	Regex *regexp.Regexp

	Path string
}

type Specification struct {
	Document *libopenapi.DocumentModel[v3.Document]
	Tree     map[string]*TreeNode
	Meta     Meta
}

type Meta struct {
	Title    string
	Version  string
	BasePath string
}

func NewSpecification(ctx context.Context, docModel *libopenapi.DocumentModel[v3.Document]) (*Specification, error) {
	// Calculate the path prefix based on the server urls
	basePath := ""
	if len(docModel.Model.Servers) > 0 {
		u, err := url.Parse(docModel.Model.Servers[0].URL)
		if err != nil {
			return nil, err
		}

		basePath = u.Path
	}

	// Create a tree for each operation
	tree := map[string]*TreeNode{}
	for _, method := range operations {
		tree[method] = &TreeNode{}
	}

	// Create the specification
	spec := &Specification{
		Document: docModel,
		Tree:     tree,
		Meta: Meta{
			Title:    docModel.Model.Info.Title,
			Version:  docModel.Model.Info.Version,
			BasePath: basePath,
		},
	}

	// Build the path tree for each operation
	for _, method := range operations {
		err := spec.buildPathTreeForMethod(ctx, docModel, method)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func (s *Specification) MatchPath(method string, p string) (*string, bool) {
	// Check if the path starts with the base path
	if !strings.HasPrefix(p, s.Meta.BasePath) {
		return nil, false
	}

	// Remove the base path from the path
	p = strings.TrimPrefix(p, s.Meta.BasePath)

	// Split the path into parts
	pathStrParts := splitPath(p)

	// Start at the root of the tree
	currentNode := s.Tree[method]
	if currentNode == nil {
		return nil, false
	}

	// Traverse the tree
	for i, part := range pathStrParts {
		isLast := i == len(pathStrParts)-1

		// TODO: Solve edge case where the path matches multiple children

		child, ok := currentNode.Children[part]
		if !ok {
			// If no child is found, check if there is a parameter that matches
			for _, c := range currentNode.Children {
				if !c.IsParameter {
					continue
				}

				if c.Regex.MatchString(part) {
					child = c

					break
				}
			}

			// If no child is found with a matching parameter, return false
			if child == nil {
				return nil, false
			}
		}

		// If an exact match of the param placeholder, return false
		// since this doesn't count as a match
		if ok && child.IsParameter {
			return nil, false
		}

		// If this is the last part of the path and the child cannot be a leaf, return false
		if isLast && !child.CanBeLeaf {
			return nil, false
		}

		currentNode = child
	}

	pathWithBase := path.Join(s.Meta.BasePath, currentNode.Path)

	return &pathWithBase, true
}

func (s *Specification) buildPathTreeForMethod(ctx context.Context, docModel *libopenapi.DocumentModel[v3.Document], method string) error {
	for pathItem := range orderedmap.Iterate(ctx, docModel.Model.Paths.PathItems) {
		// Skip the path item if the operation is not present
		if !isOperationInPathItem(pathItem.Value(), method) {
			continue
		}

		// Split the path into parts
		pathStrParts := splitPath(pathItem.Key())

		// Start at the root of the tree
		currentNode := s.Tree[method]

		// For each part of the path, create a child node
		for i, part := range pathStrParts {
			// Create the children map if it doesn't exist
			if currentNode.Children == nil {
				currentNode.Children = map[string]*TreeNode{}
			}

			// Create the child node if it doesn't exist yet
			if _, ok := currentNode.Children[part]; !ok {
				currentNode.Children[part] = &TreeNode{}
			}

			// If this is the last part of the path, mark it as a leaf
			isLastPart := i == len(pathStrParts)-1
			if isLastPart {
				currentNode.Children[part].CanBeLeaf = true
				currentNode.Children[part].Path = pathItem.Key()
			}

			// If this part is a parameter, mark it as such
			isParam := strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}")
			if isParam {
				currentNode.Children[part].IsParameter = true
			}

			// If the part is a parameter, create a regex for it
			if currentNode.Children[part].IsParameter {
				params := getPathParametersForOperationStr(method, pathItem.Value())

				re, err := makeRegexFromPath(part, params)
				if err != nil {
					return err
				}

				currentNode.Children[part].Regex = re
			}

			currentNode = currentNode.Children[part]
		}
	}

	return nil
}
