package graph

import (
	"context"
)

type DB interface {
	ListNodes(ctx context.Context) Nodes
	GetNode(ctx context.Context, guid GUID) Node
	AddNode(ctx context.Context, v Node)
	AddEdge(ctx context.Context, e Edge)
	GetGraph(ctx context.Context, guid GUID) Graph
}
