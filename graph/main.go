package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap
var graph = Graph{}

func main() {
	ctx := context.Background()
	startWatches(ctx)
	startHttpServer()
	<-ctx.Done()
}

func startHttpServer() {
	http.HandleFunc("/api/v1/graph", func(w http.ResponseWriter, r *http.Request) {
		marshal, err := json.Marshal(graph)
		if err != nil {
			panic(err)
		}
		_, err = w.Write(marshal)
		if err != nil {
			panic(err)
		}
	})
	addr := ":5678"
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()
	log.WithFields(log.Fields{"addr": addr}).Info("started")
}

func startWatches(ctx context.Context) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		panic(err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	resources, err := client.Discovery().ServerPreferredResources()
	if err != nil {
		panic(err)
	}
	d, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	for _, list := range resources {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			panic(err)
		}
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
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	log.WithField("subject", subject).Info("started watch")
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.ResultChan():
			if ok && event.Type == watch.Added {
				obj := event.Object.(*unstructured.Unstructured)
				y := Vertex(obj.GetClusterName() + "/" + obj.GetNamespace() + "/" + kind + "/" + obj.GetName())
				graph.AddVertex(y)
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
						x := Vertex(parts[0] + "/" + parts[1] + "/" + parts[2] + "/" + parts[3])
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
