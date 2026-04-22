package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "poof",
		Short: "Temporary directories that disappear after their TTL expires",
		Long:  "poof creates temporary directories with a configurable TTL.\nDirectories are automatically cleaned up via lazy sweep on every invocation.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newCmd())
	root.AddCommand(lsCmd())
	root.AddCommand(rmCmd())
	root.AddCommand(extendCmd())
	root.AddCommand(cleanCmd())
	root.AddCommand(gcCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
