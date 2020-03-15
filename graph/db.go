package graph

import (
	"context"
)

type DB interface {
	GetGraph(ctx context.Context, guid GUID) Graph
	AddNode(ctx context.Context, v Node)
	AddEdge(ctx context.Context, e Edge)
}
