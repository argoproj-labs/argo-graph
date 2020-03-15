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

var ServerCommand = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		startWatchingClusters(ctx)
		startHttpServer()
		<-ctx.Done()
	},
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

func startHttpServer() {
	http.HandleFunc("/api/v1/graph", func(w http.ResponseWriter, r *http.Request) {
		marshal, err := json.Marshal(graph)
		checkError(err)
		_, err = w.Write(marshal)
		checkError(err)
	})
	http.Handle("/", Server)
	addr := ":5678"
	go func() {
		checkError(http.ListenAndServe(addr, nil))
	}()
	log.WithFields(log.Fields{"addr": addr}).Info("started server")
}

func startWatchingClusters(ctx context.Context) {
	for clusterName, c := range getClusterConfigs() {
		startWatchingCluster(ctx, clusterName, c)
	}
}

func startWatchingCluster(ctx context.Context, clusterName string, config *rest.Config) {
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
			go watchResources(ctx, d.Resource(subject), subject, r.Name)
		}
	}
}

func watchResources(ctx context.Context, resource dynamic.ResourceInterface, subject schema.GroupVersionResource, kind string) {
	w, err := resource.Watch(metav1.ListOptions{LabelSelector: "argoproj.io/vertex"})
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
			if ok && event.Type == watch.Added {
				obj := event.Object.(*unstructured.Unstructured)
				y := NewGUID(obj.GetClusterName(), obj.GetNamespace(), kind, obj.GetName())
				label, ok := obj.GetAnnotations()["argoproj.io/vertex-label"]
				if !ok {
					label = obj.GetName()
				}
				graph.AddVertex(Vertex{GUID: y, Label: label})
				edges, ok := obj.GetAnnotations()["argoproj.io/edges"]
				if ok {
					for _, id := range strings.Split(edges, ",") {
						parts := strings.SplitN(id, "/", 4)
						if len(parts) != 4 {
							log.WithFields(log.Fields{"y": y, "x": id}).Errorf("expected 4 fields")
							continue
						}
						if parts[0] == "" {
							parts[0] = obj.GetClusterName()
						}
						if parts[1] == "" {
							parts[1] = obj.GetNamespace()
						}
						if parts[2] == "" {
							parts[2] = kind
						}
						x := NewGUID(parts[0], parts[1], parts[2], parts[3])
						e := Edge{x, y}
						graph.AddEdge(e)
						log.Infof("%v", e)
					}
				} else {
					log.WithField("y", y).Info("no inbound edges")
				}
			}
		}
	}
}
