package graph

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const namespace = "argo-graph"
const clustersSecretName = "clusters"

type RestConfig struct {
	Host        string
	APIPath     string
	Username    string
	Password    string
	BearerToken string
	rest.TLSClientConfig
	UserAgent          string
	DisableCompression bool
	QPS                float32
	Burst              int
	Timeout            time.Duration
}

func (c RestConfig) RestConfig() *rest.Config {
	return &rest.Config{
		Host:               c.Host,
		APIPath:            c.APIPath,
		Username:           c.Username,
		Password:           c.Password,
		BearerToken:        c.BearerToken,
		TLSClientConfig:    c.TLSClientConfig,
		UserAgent:          c.UserAgent,
		DisableCompression: c.DisableCompression,
		QPS:                c.QPS,
		Burst:              c.Burst,
		Timeout:            c.Timeout,
	}
}

func getRestConfig() *rest.Config {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func getKubernetes() kubernetes.Interface {
	config, err := kubernetes.NewForConfig(getRestConfig())
	checkError(err)
	return config
}
