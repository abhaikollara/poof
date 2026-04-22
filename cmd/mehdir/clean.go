package main

import (
	"fmt"
	"os"
	"time"

	"abhai.dev/mehdir/internal/registry"
	"abhai.dev/mehdir/internal/sweep"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Force a sweep of expired entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(false, false, func(reg *registry.Registry) error {
				before := len(reg.Entries)
				sweep.Run(reg, time.Now().UTC())
				removed := before - len(reg.Entries)
				if removed == 0 {
					fmt.Fprintln(os.Stderr, "mehdir: nothing to sweep")
				} else {
					fmt.Fprintf(os.Stderr, "mehdir: swept %d expired entries\n", removed)
				}
				return nil
			})
		},
	}
}
