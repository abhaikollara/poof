package main

import (
	"fmt"
	"os"

	"abhai.dev/mehdir/internal/registry"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or update configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(false, false, func(reg *registry.Registry) error {
				fmt.Fprintf(os.Stderr, "prefix: %s\n", reg.GetPrefix())
				return nil
			})
		},
	}

	cmd.AddCommand(configPrefixCmd())
	return cmd
}

func configPrefixCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prefix [VALUE]",
		Short: "Get or set the auto-generated directory prefix",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(false, false, func(reg *registry.Registry) error {
				if len(args) == 0 {
					fmt.Fprintln(os.Stdout, reg.GetPrefix())
					return nil
				}
				reg.Prefix = args[0]
				fmt.Fprintf(os.Stderr, "mehdir: prefix set to %q\n", args[0])
				return nil
			})
		},
	}
}
