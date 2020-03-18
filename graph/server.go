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
	ServerCommand.Flags().BoolVar(&dropSchema, "--drop-schema", false, "Drop the database's schema on start")
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
	http.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		marshal, err := json.Marshal(db.ListNodes(ctx))
		checkError(err)
		_, err = w.Write(marshal)
		checkError(err)
	})
	{
		pattern := "/api/v1/nodes/"
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			guid := GUID(strings.TrimPrefix(r.URL.Path, pattern))
			marshal, err := json.Marshal(db.GetNode(ctx, guid))
			checkError(err)
			_, err = w.Write(marshal)
			checkError(err)
		})
	}
	{
		pattern := "/api/v1/graph/"
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			guid := GUID(strings.TrimPrefix(r.URL.Path, pattern))
			marshal, err := json.Marshal(db.GetGraph(ctx, guid))
			checkError(err)
			_, err = w.Write(marshal)
			checkError(err)
		})
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
			if ok && event.Type != watch.Deleted {
				obj := event.Object.(*unstructured.Unstructured)
				x := NewGUID(clusterName, obj.GetNamespace(), kind, obj.GetName())
				label, ok := obj.GetAnnotations()["graph.argoproj.io/node-label"]
				if !ok {
					label = obj.GetName()
				}
				db.AddNode(ctx, Node{GUID: x, Label: label})
				edges, ok := obj.GetAnnotations()["graph.argoproj.io/edges"]
				if ok {
					for _, id := range strings.Split(edges, ",") {
						parts := strings.SplitN(id, "/", 4)
						if len(parts) != 4 {
							log.WithFields(log.Fields{"x": x, "y": id}).Errorf("expected 4 fields")
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
						e := Edge{x, y}
						db.AddEdge(ctx, e)
						log.Infof("%v", e)
					}
				} else {
					log.WithField("x", x).Info("no inbound edges")
				}
			}
		}
	}
}
