package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DB struct {
	conn *grpc.ClientConn
	dg   *dgo.Dgraph
}

func (d *DB) getUID(ctx context.Context, xid GUID) string {
	q := `query Me($xid: string){
	me(func: eq(xid, $xid)) {
		uid
    }
}`
	resp, err := d.dg.NewTxn().QueryWithVars(ctx, q, map[string]string{"$xid": string(xid)})
	checkError(err)
	type Root struct {
		Me []Resource `json:"me"`
	}
	var r Root
	checkError(json.Unmarshal(resp.Json, &r))
	if len(r.Me) > 0 {
		return r.Me[0].UID
	}
	return ""
}

func (d *DB) AddVertex(ctx context.Context, v Vertex) {

	uid := d.getUID(ctx, v.GUID)

	if uid == "" {
		log.WithField("vertex", v).Debug("AddVertex")
		_, err := d.dg.NewTxn().Mutate(ctx, &api.Mutation{CommitNow: true, SetJson: []byte(fmt.Sprintf(`{"set": [{"xid": "%s", "label": "%s"}]}`, v.GUID, v.Label))})
		checkError(err)
	} else if v.Label != "" {
		log.WithField("vertex", v).Debug("AddVertex (upsert)")
		_, err := d.dg.NewTxn().Mutate(ctx, &api.Mutation{CommitNow: true, SetJson: []byte(fmt.Sprintf(`{"set": [{"uid": "%s", "xid": "%s", "label": "%s"}]}`, uid, v.GUID, v.Label))})
		checkError(err)
	}
}

func (d *DB) AddEdge(ctx context.Context, e Edge) {
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

func (d *DB) GetGraph(ctx context.Context) Graph {
	resp, err := d.dg.NewTxn().QueryWithVars(ctx, `query Me(){
	me(func: has(xid))  {
        xid
        label
        follows {
           xid
        }
    }
}`, map[string]string{})
	checkError(err)
	type Root struct {
		Me []Resource `json:"me"`
	}
	var r Root
	checkError(json.Unmarshal(resp.Json, &r))
	g := Graph{}
	for _, r := range r.Me {
		g.AddVertex(Vertex{GUID: GUID(r.XID), Label: r.Label})
		for _, y := range r.Follows {
			g.AddEdge(Edge{GUID(r.XID), GUID(y.XID)})
		}
	}
	return g
}

var db = NewDB()

type Resource struct {
	UID     string     `json:"uid,omitempty"`
	XID     string     `json:"xid,omitempty"`
	Label   string     `json:"label,omitempty"`
	Follows []Resource `json:"follows,omitempty"`
}

func NewDB() *DB {
	log.Info("creating database connection")
	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	checkError(err)
	dg := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	ctx := context.Background()
	log.Info("dropping database schema")
	err = dg.Alter(ctx, &api.Operation{DropOp: api.Operation_ALL})
	checkError(err)
	log.Info("creating database schema")
	err = dg.Alter(ctx, &api.Operation{Schema: `
	xid: string @index(exact) @upsert .
	label: string .
    follows: [uid] .

type Resource {
  xid
  label
  follows
 }
`})
	checkError(err)
	log.Info("database connection ready")
	return &DB{conn: conn, dg: dg}
}
