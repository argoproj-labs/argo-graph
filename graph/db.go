package graph

import (
	"context"
)

type DB interface {
	GetGraph(ctx context.Context, guid GUID) Graph
	AddNode(ctx context.Context, v Node)
	ListNodes(ctx context.Context) Nodes
	AddEdge(ctx context.Context, e Edge)
}
