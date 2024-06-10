package swagger

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type Specification struct {
	Document *libopenapi.DocumentModel[v3.Document]
	Tree     map[string]*Node
	Meta     Meta
}

type Meta struct {
	Title    string
	Version  string
	BasePath string
}

type Node struct {
	Children map[string]*Node

	IsParameter bool
	CanBeLeaf   bool

	Regex *regexp.Regexp

	Path string
}

func (n *Node) MatchParam(part string) bool {
	if !n.IsParameter {
		return false
	}

	return n.Regex.MatchString(part)
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
	tree := map[string]*Node{}
	for _, method := range operations {
		tree[method] = &Node{}
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

func (s *Specification) MatchPath(method string, p string) (*Node, bool) {
	// Check if the path starts with the base path
	if !strings.HasPrefix(p, s.Meta.BasePath) {
		return nil, false
	}

	// Remove the base path from the path
	p = strings.TrimPrefix(p, s.Meta.BasePath)

	// Split the path into parts
	pathStrParts := splitPath(p)

	// Start at the root of the tree
	tree := s.Tree[method]
	if tree == nil {
		return nil, false
	}

	// Perform a depth-first search on the tree
	return s.dfs(tree, pathStrParts)
}

type stackNode struct {
	Node      *Node
	PartIndex int
}

func (s *Specification) dfs(currentNode *Node, pathStrParts []string) (*Node, bool) {
	// Create a stack for the depth-first search with the root node
	stack := []stackNode{{Node: currentNode, PartIndex: 0}}
	pathLen := len(pathStrParts)

	for len(stack) > 0 {
		// Pop the top of the stack
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Get the current node and part index from the stack
		currentNode, partIndex := top.Node, top.PartIndex

		// If the current node is nil, continue to the next iteration
		if currentNode == nil {
			continue
		}

		// If this is the last part of the path and the current node can be a leaf, return the path
		isLastPart := partIndex == pathLen
		if isLastPart {
			if currentNode.CanBeLeaf {
				return currentNode, true
			}

			continue
		}

		part := pathStrParts[partIndex]

		// Check for an exact match
		if child, ok := currentNode.Children[part]; ok {
			// Exact match not allowed for parameters
			if !child.IsParameter {
				stack = append(stack, stackNode{Node: child, PartIndex: partIndex + 1})

				continue
			}
		}

		// If no exact match was found, check for parameter matches
		potentialMatches := []*Node{}
		for _, child := range currentNode.Children {
			if child.IsParameter && child.MatchParam(part) {
				potentialMatches = append(potentialMatches, child)
			}
		}

		// Push the potential matches to the stack
		for _, match := range potentialMatches {
			stack = append(stack, stackNode{Node: match, PartIndex: partIndex + 1})
		}
	}

	return nil, false
}

func (s *Specification) buildPathTreeForMethod(
	ctx context.Context,
	docModel *libopenapi.DocumentModel[v3.Document],
	method string,
) error {
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
				currentNode.Children = map[string]*Node{}
			}

			// Create the child node if it doesn't exist yet
			if _, ok := currentNode.Children[part]; !ok {
				currentNode.Children[part] = &Node{}
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
