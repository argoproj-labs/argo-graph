package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type dgraphDB struct {
	conn *grpc.ClientConn
	dg   *dgo.Dgraph
}

func (d *dgraphDB) GetNode(ctx context.Context, guid GUID) Node {
	resp, err := d.dg.NewReadOnlyTxn().Query(ctx, `query Me(){
	resources(func: eq(xid, "`+string(guid)+`")) {
        xid
        label
        phase
    }
}`)
	checkError(err)
	var r Root
	checkError(json.Unmarshal(resp.Json, &r))
	for _, r := range r.Resources {
		return Node{GUID: GUID(r.XID), Label: r.Label, Phase: r.Phase}
	}
	return Node{}
}

type Root struct {
	Resources []Resource `json:"resources"`
}

func (d *dgraphDB) ListNodes(ctx context.Context) Nodes {
	resp, err := d.dg.NewReadOnlyTxn().Query(ctx, `query Me(){
	resources(func: has(xid)) {
        xid
        label
		phase
    }
}`)
	checkError(err)
	var r Root
	checkError(json.Unmarshal(resp.Json, &r))
	nodes := make(Nodes, len(r.Resources))
	for i, r := range r.Resources {
		nodes[i] = Node{GUID: GUID(r.XID), Label: r.Label, Phase: r.Phase}
	}
	return nodes
}

func (d *dgraphDB) getUID(ctx context.Context, xid GUID) string {
	resp, err := d.dg.NewTxn().QueryWithVars(ctx, `query Me($xid: string){
	resources(func: eq(xid, $xid)) {
		uid
    }
}`, map[string]string{"$xid": string(xid)})
	checkError(err)
	var r Root
	checkError(json.Unmarshal(resp.Json, &r))
	if len(r.Resources) > 0 {
		return r.Resources[0].UID
	}
	return ""
}

func (d *dgraphDB) AddNode(ctx context.Context, v Node) {
	uid := d.getUID(ctx, v.GUID)
	if uid == "" {
		log.WithField("node", v).Debug("AddNode")
		_, err := d.dg.NewTxn().Mutate(ctx, &api.Mutation{CommitNow: true, SetJson: []byte(fmt.Sprintf(`{"set": [{"xid": "%s", "label": "%s", "phase": "%s"}]}`, v.GUID, v.Label, v.Phase))})
		checkError(err)
	} else if v.Label != "" || v.Phase != "" {
		log.WithField("node", v).Debug("AddNode (upsert)")
		_, err := d.dg.NewTxn().Mutate(ctx, &api.Mutation{CommitNow: true, SetJson: []byte(fmt.Sprintf(`{"set": [{"uid": "%s", "xid": "%s", "label": "%s", "phase": "%s"}]}`, uid, v.GUID, v.Label, v.Phase))})
		checkError(err)
	}
}

func (d *dgraphDB) AddEdge(ctx context.Context, e Edge) {
	log.WithField("edge", e).Debug("AddEdge")
	_, err := d.dg.NewTxn().Mutate(ctx, &api.Mutation{CommitNow: true, SetJson: []byte(fmt.Sprintf(`{
  "set":[
    {
      "uid": "%s",
      "follows": {
        "uid": "%s"
      }
    }
  ]
}`, d.getUID(ctx, e.X), d.getUID(ctx, e.Y)))})
	checkError(err)
}

func add(g *Graph, rs []Resource) {
	for _, r := range rs {
		g.AddNode(Node{GUID: GUID(r.XID), Label: r.Label, Phase: r.Phase})
		for _, y := range r.Follows {
			g.AddEdge(Edge{GUID(r.XID), GUID(y.XID)})
		}
		for _, y := range r.Followers {
			g.AddEdge(Edge{GUID(y.XID), GUID(r.XID)})
		}
		add(g, r.Follows)
		add(g, r.Followers)
	}
}

func (d *dgraphDB) GetGraph(ctx context.Context, guid GUID) Graph {
	resp, err := d.dg.NewTxn().Query(ctx, fmt.Sprintf(`query Me(){
	resources(func: eq(xid, "%s")) @recurse(depth: 10){
        xid
        label
        phase
        follows
        followers: ~follows
    }
}`, guid))
	checkError(err)
	var r Root
	checkError(json.Unmarshal(resp.Json, &r))
	g := &Graph{}
	add(g, r.Resources)
	return *g
}

type Resource struct {
	UID       string     `json:"uid,omitempty"`
	XID       string     `json:"xid,omitempty"`
	Label     string     `json:"label,omitempty"`
	Phase     string     `json:"phase,omitempty"`
	Follows   []Resource `json:"follows,omitempty"`
	Followers []Resource `json:"followers,omitempty"`
}

func NewDB(dropSchema bool) DB {
	log.Info("creating database connection")
	conn, err := grpc.Dial(os.Getenv("DGRAPH_HOST")+":9080", grpc.WithInsecure())
	checkError(err)
	dg := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	ctx := context.Background()
	if dropSchema {
		log.Info("dropping database schema")
		err = dg.Alter(ctx, &api.Operation{DropOp: api.Operation_ALL})
	}
	checkError(err)
	log.Info("creating database schema")
	err = dg.Alter(ctx, &api.Operation{Schema: `
	xid: string @index(exact) @upsert .
	label: string .
	phase: string .
    follows: [uid] @reverse .

type Resource {
  xid
  label
  phase
  follows
 }
`})
	checkError(err)
	log.Info("database connection ready")
	return &dgraphDB{conn: conn, dg: dg}
}
