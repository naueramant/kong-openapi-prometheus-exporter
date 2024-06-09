package swagger

import (
	"fmt"
	"os"
	"path"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/google/uuid"
)

func Visualize(spec *Specification, outDir string) error {
	if err := os.Mkdir(outDir, 0o755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	for _, method := range operations {
		g := graphviz.New()
		graph, err := g.Graph()
		if err != nil {
			return err
		}

		rootNode, err := graph.CreateNode(method)
		if err != nil {
			return err
		}

		buildGraph(graph, spec.Tree[method], rootNode)

		if err := g.RenderFilename(graph, graphviz.PNG, path.Join(
			outDir,
			fmt.Sprintf("%s.png", method),
		)); err != nil {
			return err
		}
	}

	return nil
}

func buildGraph(graph *cgraph.Graph, treeNode *Node, graphNode *cgraph.Node) error {
	for part, child := range treeNode.Children {
		id := uuid.New().String()

		childNode, err := graph.CreateNode(part + " " + id)
		if err != nil {
			return err
		}

		if child.IsParameter {
			childNode.SetLabel(fmt.Sprintf("%s\n%s", part, child.Regex.String()))
			childNode.SetColor("green")
		} else {
			childNode.SetLabel(part)
		}

		if child.CanBeLeaf {
			childNode.SetFontColor("red")
		}

		if _, err := graph.CreateEdge("", graphNode, childNode); err != nil {
			return err
		}

		buildGraph(graph, child, childNode)
	}

	return nil
}
