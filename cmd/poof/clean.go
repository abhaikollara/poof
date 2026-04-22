package main

import (
	"fmt"
	"os"

	"abhai.dev/poof/internal/registry"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Force a sweep of expired entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(true, false, func(reg *registry.Registry) error {
				fmt.Fprintln(os.Stderr, "poof: sweep complete")
				return nil
			})
		},
	}
}
