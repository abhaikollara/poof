package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"abhai.dev/poof/internal/registry"
	"abhai.dev/poof/internal/ttl"
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
				fmt.Fprintf(os.Stderr, "poof: resolving path: %v\n", err)
				os.Exit(exitUserError)
			}

			dur, err := ttl.Parse(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "poof: %v\n", err)
				os.Exit(exitUserError)
			}

			return withRegistry(true, false, func(reg *registry.Registry) error {
				_, entry := reg.FindByPath(path)
				if entry == nil {
					fmt.Fprintf(os.Stderr, "poof: %q not found\n", path)
					os.Exit(exitUserError)
				}

				entry.ExpiresAt = time.Now().UTC().Add(dur)
				fmt.Fprintf(os.Stderr, "poof: %s now expires at %s\n", path, entry.ExpiresAt.Local().Format(time.RFC3339))
				return nil
			})
		},
	}
}
