package graph

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var ClusterCommand = &cobra.Command{
	Use: "cluster",
	Run: func(c *cobra.Command, args []string) {
		c.HelpFunc()(c, args)
	},
}

func init() {
	ClusterCommand.AddCommand(&cobra.Command{
		Use: "add [CONTEXT_NAME...]",
		Run: func(cmd *cobra.Command, args []string) {
			for _, contextName := range args {
				startingConfig, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
				checkError(err)
				configOverrides := &clientcmd.ConfigOverrides{Context: *startingConfig.Contexts[contextName]}
				restConfig := restConfig(clientcmd.NewDefaultClientConfig(*startingConfig, configOverrides))
				secrets := getKubernetes().CoreV1().Secrets(namespace)
				secret, err := secrets.Get(clustersSecretName, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					secret, err = secrets.Create(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "clusters"}})
					checkError(err)
				}
				if secret.StringData == nil {
					secret.StringData = map[string]string{}
				}
				marshal, err := json.MarshalIndent(restConfig, "", "  ")
				checkError(err)
				log.WithField("marshal", string(marshal)).Debug()
				secret.StringData[contextName] = string(marshal)
				_, err = secrets.Update(secret)
				checkError(err)
				fmt.Printf(`added cluster context "%s"\n`, contextName)
			}
		},
	})
}

func restConfig(config clientcmd.ClientConfig) RestConfig {
	c, err := config.ClientConfig()
	checkError(err)
	log.WithField("config", config).Debug()
	return RestConfig{
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
