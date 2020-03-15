package graph

import (
	"context"
)

type DB interface {
	GetGraph(ctx context.Context) Graph
	AddNode(ctx context.Context, v Node)
	AddEdge(ctx context.Context, e Edge)
}
