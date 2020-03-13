package main

import "fmt"

// https: //en.wikipedia.org/wiki/Graph_(discrete_mathematics)
type Vertex string
type Edge struct {
	X, Y Vertex
}

func (e Edge) String() string {
	return fmt.Sprintf("%s -> %s", e.X, e.Y)
}

type Vertices []Vertex
type Edges []Edge
type Graph struct {
	Vertices `json:"vertices"`
	Edges    `json:"edge"`
}

func (g Graph) AddVertex(v Vertex) {
	g.Vertices = append(g.Vertices, v)
}

func (g Graph) AddEdge(e Edge) {
	g.Edges = append(g.Edges, e)
}
