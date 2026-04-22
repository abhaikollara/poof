package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableTraverseRunHooks = true

	root := &cobra.Command{
		Use:   "poof",
		Short: "Temporary directories that disappear after their TTL expires",
		Long:  "poof creates temporary directories with a configurable TTL.\nDirectories are automatically cleaned up by a background daemon.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "poof: %v\n", err)
		os.Exit(exitUserError)
		return nil
	})

	root.AddCommand(newCmd())
	root.AddCommand(lsCmd())
	root.AddCommand(rmCmd())
	root.AddCommand(extendCmd())
	root.AddCommand(cleanCmd())
	root.AddCommand(gcCmd())
	root.AddCommand(daemonCmd())

	if err := root.Execute(); err != nil {
		os.Exit(exitUserError)
	}
}
