package main

import (
	"fmt"
	"os"

	"abhai.dev/mehdir/internal/registry"
	"abhai.dev/mehdir/internal/sweep"
	"github.com/spf13/cobra"
)

func gcCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gc",
		Short: "Remove registry entries whose directories no longer exist",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(true, false, func(reg *registry.Registry) error {
				removed := sweep.GC(reg)
				if removed == 0 {
					fmt.Fprintln(os.Stderr, "mehdir: no orphaned entries")
				} else {
					fmt.Fprintf(os.Stderr, "mehdir: removed %d orphaned entries\n", removed)
				}
				return nil
			})
		},
	}
}
