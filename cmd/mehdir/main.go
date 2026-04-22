package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableTraverseRunHooks = true

	root := &cobra.Command{
		Use:   "mehdir",
		Short: "Temporary directories that disappear after their TTL expires",
		Long:  "mehdir creates temporary directories with a configurable TTL.\nDirectories are automatically cleaned up by a background daemon.",
		SilenceUsage: true,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return newCmd().RunE(cmd, args)
		},
	}

	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
		os.Exit(exitUserError)
		return nil
	})

	root.AddCommand(newCmd())
	root.AddCommand(addCmd())
	root.AddCommand(lsCmd())
	root.AddCommand(rmCmd())
	root.AddCommand(extendCmd())
	root.AddCommand(cleanCmd())
	root.AddCommand(cleanslateCmd())
	root.AddCommand(configCmd())
	root.AddCommand(gcCmd())
	root.AddCommand(daemonCmd())

	if err := root.Execute(); err != nil {
		os.Exit(exitUserError)
	}
}
