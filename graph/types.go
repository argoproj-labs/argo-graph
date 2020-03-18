package graph

import (
	"fmt"
	"strings"
)

// https: //en.wikipedia.org/wiki/Graph_(discrete_mathematics)
type GUID string

func NewGUID(cluster, namespace, kind, name string) GUID {
	return GUID(cluster + "/" + namespace + "/" + kind + "/" + name)
}

type Node struct {
	GUID  GUID   `json:"guid"`
	Label string `json:"label"`
	Phase string `json:"phase,omitempty"`
}

func (v Node) GetKind() string {
	return v.parts()[2]
}

func (v Node) parts() []string {
	return strings.SplitN(string(v.GUID), "/", 4)
}

func (v Node) IsZero() bool {
	return v.GUID == ""
}

type Edge struct {
	X GUID `json:"x"`
	Y GUID `json:"y"`
}

func (e Edge) String() string {
	return fmt.Sprintf("%s -> %s", e.X, e.Y)
}

type Nodes []Node
type Edges []Edge
type Graph struct {
	Nodes Nodes `json:"nodes"`
	Edges Edges `json:"edges"`
}

func (g *Graph) AddNode(v Node) {
	g.Nodes = append(g.Nodes, v)
}

func (g *Graph) AddEdge(e Edge) {
	g.Edges = append(g.Edges, e)
}
