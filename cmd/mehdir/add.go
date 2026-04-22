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

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <PATH> <TTL>",
		Short: "Track an existing directory for automatic deletion",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: resolving path: %v\n", err)
				os.Exit(exitUserError)
			}

			info, err := os.Stat(targetPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %s does not exist\n", targetPath)
				os.Exit(exitUserError)
			}
			if !info.IsDir() {
				fmt.Fprintf(os.Stderr, "mehdir: %s is not a directory\n", targetPath)
				os.Exit(exitUserError)
			}

			dur, err := ttl.Parse(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitUserError)
			}

			err = withRegistry(true, false, func(reg *registry.Registry) error {
				if _, existing := reg.FindByPath(targetPath); existing != nil {
					fmt.Fprintf(os.Stderr, "mehdir: %s is already tracked\n", targetPath)
					os.Exit(exitUserError)
				}

				reg.AddAllowedPrefix(filepath.Dir(targetPath))

				now := time.Now().UTC()
				reg.Entries = append(reg.Entries, registry.Entry{
					Path:      targetPath,
					CreatedAt: now,
					ExpiresAt: now.Add(dur),
				})
				return nil
			})
			if err != nil {
				return err
			}

			ensureDaemon()

			fmt.Fprintf(os.Stderr, "mehdir: tracking %s (expires in %s)\n", targetPath, args[1])
			return nil
		},
	}
}
