package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap
var graph = Graph{}

func main() {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		panic(err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clusterName := ""
	_, controller := cache.NewInformer(&cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (object runtime.Object, err error) {
			opts.LabelSelector = "argoproj.io/vertex"
			return client.CoreV1().Pods("").List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (w watch.Interface, err error) {
			opts.LabelSelector = "argoproj.io/vertex"
			return client.CoreV1().Pods("").Watch(opts)
		},
	}, &v1.Pod{}, time.Second*30, cache.ResourceEventHandlerFuncs{
		AddFunc: func(iface interface{}) {
			obj := iface.(*v1.Pod)
			y := Vertex(obj.GetClusterName() + "/" + obj.GetNamespace() + "/" + obj.Kind + "/" + obj.GetName())
			graph.AddVertex(y)
			edges, ok := obj.GetAnnotations()["argoproj.io/edges"]
			if ok {
				for _, id := range strings.Split(edges, ",") {
					parts := strings.SplitN(id, "/", 4)
					if parts[0] == "" {
						parts[0] = obj.GetClusterName()
					}
					if parts[1] == "" {
						parts[1] = obj.GetNamespace()
					}
					if parts[2] == "" {
						parts[2] = obj.Kind
					}
					x := Vertex(parts[0] + "/" + parts[1] + "/" + parts[2] + "/" + parts[3])
					e := Edge{x, y}
					graph.AddEdge(e)
					log.Infof("%v", e)
				}
			} else {
				log.WithField("y", y).Info("no inbound edges")
			}
		},
	})
	stop := make(chan struct{})
	go controller.Run(stop)

	http.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		marshal, err := json.Marshal(graph)
		if err != nil {
			panic(err)
		}
		_, err = w.Write(marshal)
		if err != nil {
			panic(err)
		}
	})
	addr := ":2746"
	go func() {
		err = http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()
	log.WithFields(log.Fields{"clusterName": clusterName, "addr": addr}).Info("started")
	<-stop
}
