package main

import (
	"fmt"
	"os"
	"path/filepath"

	"abhai.dev/mehdir/internal/registry"
	"github.com/spf13/cobra"
)

func rmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <PATH>",
		Short: "Delete a tracked directory immediately",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: resolving path: %v\n", err)
				os.Exit(exitUserError)
			}

			return withRegistry(true, false, func(reg *registry.Registry) error {
				i, entry := reg.FindByPath(path)
				if i == -1 {
					fmt.Fprintf(os.Stderr, "mehdir: %q not found\n", path)
					os.Exit(exitUserError)
				}

				if err := registry.SafeToDelete(entry.Path, reg.AllowedPrefixes); err != nil {
					fmt.Fprintf(os.Stderr, "mehdir: safety check failed: %v\n", err)
					os.Exit(exitInternalError)
				}

				if err := os.RemoveAll(entry.Path); err != nil {
					fmt.Fprintf(os.Stderr, "mehdir: removing %s: %v\n", entry.Path, err)
				}

				reg.RemoveIndex(i)
				fmt.Fprintf(os.Stderr, "mehdir: removed %s\n", entry.Path)
				return nil
			})
		},
	}
}
