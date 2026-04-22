package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"abhai.dev/mehdir/internal/registry"
	"abhai.dev/mehdir/internal/ttl"
	"github.com/spf13/cobra"
)

func extendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extend <PATH> <TTL>",
		Short: "Replace the expiry with now + duration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: resolving path: %v\n", err)
				os.Exit(exitUserError)
			}

			dur, err := ttl.Parse(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitUserError)
			}

			return withRegistry(true, false, func(reg *registry.Registry) error {
				_, entry := reg.FindByPath(path)
				if entry == nil {
					fmt.Fprintf(os.Stderr, "mehdir: %q not found\n", path)
					os.Exit(exitUserError)
				}

				entry.ExpiresAt = time.Now().UTC().Add(dur)
				fmt.Fprintf(os.Stderr, "mehdir: %s now expires at %s\n", path, entry.ExpiresAt.Local().Format(time.RFC3339))
				return nil
			})
		},
	}
}
