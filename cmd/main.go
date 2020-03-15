package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/argoproj/argo-graph/graph"
)

func main() {
	var command = &cobra.Command{
		Use: "argo-graph",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}

	command.AddCommand(graph.ClusterCommand)
	command.AddCommand(graph.ServerCommand)

	// global log level
	var logLevel string
	command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			panic(err)
		}
		log.SetLevel(level)
	}
	command.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")

	err := command.Execute()
	if err != nil {
		panic(err)
	}
}
