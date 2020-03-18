package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var dropSchema bool

var ServerCommand = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		db := NewDB(dropSchema)
		startWatchingClusters(ctx, db)
		startHttpServer(ctx, db)
		<-ctx.Done()
	},
}

func init() {
	ServerCommand.Flags().BoolVar(&dropSchema, "drop-schema", false, "Drop the database's schema on start")
}

func getClusterConfigs() map[string]*rest.Config {
	secrets, err := getKubernetes().CoreV1().Secrets(namespace).Get(clustersSecretName, metav1.GetOptions{})
	checkError(err)
	configs := map[string]*rest.Config{}
	for name, bytes := range secrets.Data {
		r := &RestConfig{}
		err := json.Unmarshal(bytes, r)
		checkError(err)
		configs[name] = r.RestConfig()
	}
	return configs
}

func startHttpServer(ctx context.Context, db DB) {

	jsonFunc := func(f func(r *http.Request) (int, interface{})) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, q *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					marshal, _ := json.Marshal(r)
					w.WriteHeader(500)
					_, _ = w.Write(marshal)
					log.WithFields(log.Fields{"url": q.URL, "statusCode": 500}).Error(r)
				}
			}()
			statusCode, v := f(q)
			marshal, err := json.Marshal(v)
			checkError(err)
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(marshal)
			checkError(err)
			log.WithFields(log.Fields{"url": q.URL, "statusCode": statusCode}).Debug()
		}
	}
	http.HandleFunc("/api/v1/nodes", jsonFunc(func(r *http.Request) (int, interface{}) {
		return 200, db.ListNodes(ctx)
	}))
	{
		pattern := "/api/v1/nodes/"
		http.HandleFunc(pattern, jsonFunc(func(r *http.Request) (int, interface{}) {
			guid := GUID(strings.TrimPrefix(r.URL.Path, pattern))
			node := db.GetNode(ctx, guid)
			if node.IsZero() {
				return 404, nil
			}
			return 200, node
		}))
	}
	{
		pattern := "/api/v1/graph/"
		http.HandleFunc(pattern, jsonFunc(func(r *http.Request) (int, interface{}) {
			guid := GUID(strings.TrimPrefix(r.URL.Path, pattern))
			return 200, db.GetGraph(ctx, guid)
		}))
	}
	http.Handle("/", Server)
	addr := ":5678"
	go func() {
		checkError(http.ListenAndServe(addr, nil))
	}()
	log.WithFields(log.Fields{"addr": addr}).Info("started server")
}

func startWatchingClusters(ctx context.Context, db DB) {
	for clusterName, c := range getClusterConfigs() {
		startWatchingCluster(ctx, clusterName, c, db)
	}

	log.Info("cluster watches started")
}

func startWatchingCluster(ctx context.Context, clusterName string, config *rest.Config, db DB) {
	log.WithFields(log.Fields{"clusterName": clusterName}).Info("starting watching cluster")
	client, err := kubernetes.NewForConfig(config)
	checkError(err)
	resources, err := client.Discovery().ServerPreferredResources()
	checkError(err)
	d, err := dynamic.NewForConfig(config)
	checkError(err)
	for _, list := range resources {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		checkError(err)
		for _, r := range list.APIResources {
			subject := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: r.Name}
			go watchResources(ctx, d.Resource(subject), subject, clusterName, r.Name, db)
		}
	}
}

func watchResources(ctx context.Context, resource dynamic.ResourceInterface, subject schema.GroupVersionResource, clusterName, kind string, db DB) {
	w, err := resource.Watch(metav1.ListOptions{LabelSelector: "graph.argoproj.io/node"})
	if errors.IsNotFound(err) || errors.IsMethodNotSupported(err) {
		log.WithField("subject", subject).Info(err)
		return
	}
	checkError(err)
	defer w.Stop()
	log.WithField("subject", subject).Info("started watch")
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.ResultChan():
			if ok {
				if event.Type != watch.Deleted {
					obj := event.Object.(*unstructured.Unstructured)
					node := NewGUID(clusterName, obj.GetNamespace(), kind, obj.GetName())
					log.WithField("guid", node).Debug()
					label, ok := obj.GetAnnotations()["graph.argoproj.io/node-label"]
					if !ok {
						label = obj.GetName()
					}
					phase := ""
					spec, ok := obj.Object["status"].(map[string]interface{})
					if ok {
						p, ok := spec["phase"]
						if ok {
							phase = p.(string)
						}
					}
					db.AddNode(ctx, Node{GUID: node, Label: label, Phase: phase})
					edges, ok := obj.GetAnnotations()["graph.argoproj.io/edges"]
					if ok {
						for _, id := range strings.Split(edges, ",") {
							parts := strings.SplitN(id, "/", 4)
							if len(parts) != 4 {
								log.WithFields(log.Fields{"x": node, "y": id}).Errorf("expected 4 fields")
								continue
							}
							if parts[0] == "" {
								parts[0] = clusterName
							}
							if parts[1] == "" {
								parts[1] = obj.GetNamespace()
							}
							if parts[2] == "" {
								parts[2] = kind
							}
							y := NewGUID(parts[0], parts[1], parts[2], parts[3])
							db.AddNode(ctx, Node{GUID: y})
							e := Edge{node, y}
							db.AddEdge(ctx, e)
							log.Infof("%v", e)
						}
					} else {
						log.WithField("x", node).Info("no inbound edges")
					}
				}
			} else {
				log.Warn("not ok")
			}
		}
	}
}
