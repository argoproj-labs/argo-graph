package main

import (
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
	command.AddCommand(graph.ServerCommand)
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}
