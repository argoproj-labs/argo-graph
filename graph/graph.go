package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type vertex string
type edge struct {
	a, b vertex
}

// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap

func main() {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		panic(err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	_, controller := cache.NewInformer(&cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (object runtime.Object, err error) {
			opts.LabelSelector = "argoproj.io/vertex"
			return client.CoreV1().Pods("").List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (w watch.Interface, err error) {
			opts.LabelSelector = "argoproj.io/vertex"
			return client.CoreV1().Pods("").Watch(opts)
		},
	}, &corev1.Pod{}, time.Second*30, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			o := obj.(metav1.Object)
			edges := o.GetLabels()["argoproj.io/edges"]

			log.WithField("name", o.GetName()).Info()
		},
	})
	log.Info("started")
	stop := make(chan struct{})
	go controller.Run(stop)
	<-stop
}
