package main

import "fmt"

// https: //en.wikipedia.org/wiki/Graph_(discrete_mathematics)
type GUID string
type Vertex struct {
	GUID  GUID   `json:"guid"`
	Label string `json:"label"`
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
