package main

import (
	"fmt"
	"os"

	"abhai.dev/mehdir/internal/registry"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Force a sweep of expired entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(true, false, func(reg *registry.Registry) error {
				fmt.Fprintln(os.Stderr, "mehdir: sweep complete")
				return nil
			})
		},
	}
}
