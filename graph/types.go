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

type Vertex struct {
	GUID  GUID   `json:"guid"`
	Label string `json:"label"`
}

func (v Vertex) GetKind() string {
	return v.parts()[2]
}

func (v Vertex) parts() []string {
	return strings.SplitN(string(v.GUID), "/", 4)
}

type Edge struct {
	X GUID `json:"x"`
	Y GUID `json:"y"`
}

func (e Edge) String() string {
	return fmt.Sprintf("%s -> %s", e.X, e.Y)
}

type Vertices []Vertex
type Edges []Edge
type Graph struct {
	Vertices `json:"vertices"`
	Edges    `json:"edges"`
}

func (g *Graph) AddVertex(v Vertex) {
	g.Vertices = append(g.Vertices, v)
}

func (g *Graph) AddEdge(e Edge) {
	g.Edges = append(g.Edges, e)
}
