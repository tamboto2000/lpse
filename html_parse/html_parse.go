package html_parse

import (
	"bytes"

	"golang.org/x/net/html"
)

// Node is slightly modified version of golang.org/x/net/html.Node
type Node struct {
	*html.Node
	Childs []*Node
}

// ParseBytes parse HTML bytes to marshalable node
func ParseBytes(byts []byte) (*Node, error) {
	node, err := html.Parse(bytes.NewBuffer(byts))
	if err != nil {
		return nil, err
	}

	return parseFromNode(node), nil
}

// SearchNode search a node matched with params.
// ty for HTML object type,
// data is for HTML tag name,
// key is for attribute key
// val is for attribute value with key
func (n *Node) SearchNode(ty html.NodeType, data, namespace, key, val string) *Node {
	return searchNode(ty, data, namespace, key, val, n)
}

// SearchAllNode search nodes matched with options.
// ty for HTML object type,
// data is for HTML tag name,
// key is for attribute key
// val is for attribute value with key
func (n *Node) SearchAllNode(ty html.NodeType, data, namespace, key, val string) []Node {
	return searchAllNode(ty, data, namespace, key, val, n)
}

func searchAllNode(ty html.NodeType, data, namespace, key, val string, node *Node) []Node {
	nodes := make([]Node, 0)
	if isNodeMatch(ty, data, namespace, key, val, node) {
		nodes = append(nodes, *node)
	}

	if node.Childs != nil {
		for _, child := range node.Childs {
			nodes = append(nodes, searchAllNode(ty, data, namespace, key, val, child)...)
		}
	}

	return nodes
}

func searchNode(ty html.NodeType, data, namespace, key, val string, node *Node) *Node {
	if isNodeMatch(ty, data, namespace, key, val, node) {
		return node
	}

	if node.Childs != nil {
		for _, child := range node.Childs {
			newNode := searchNode(ty, data, namespace, key, val, child)
			if newNode != nil {
				return newNode
			}
		}
	}

	return nil
}

func isNodeMatch(ty html.NodeType, data, namespace, key, val string, node *Node) bool {
	if ty < 8 && node.Type != ty {
		return false
	}

	if data != "" && node.Data != data {
		return false
	}

	if namespace != "" && node.Namespace != namespace {
		return false
	}

	if key != "" {
		found := false
		for _, attr := range node.Attr {
			if attr.Key == key {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	if val != "" {
		found := false
		for _, attr := range node.Attr {
			if attr.Val == val {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func parseFromNode(root *html.Node) *Node {
	newNode := new(Node)
	newNode.Node = root

	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if result := parseFromNode(c); result != nil {
			newNode.Childs = append(newNode.Childs, result)
		}
	}

	return newNode
}
