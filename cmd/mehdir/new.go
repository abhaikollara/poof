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

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new [PATH] [TTL]",
		Short: "Create a new temporary directory",
		Long: `Create a new temporary directory.

If PATH is given, that directory is created and tracked directly.
If only a TTL is given, an auto-named directory is created in the current directory.
If no arguments are given, an auto-named directory is created with a 1h TTL.

The auto-generated prefix defaults to "mehdir-" and can be changed with "mehdir config prefix".`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetPath string
			ttlStr := "1h"
			autoName := true

			switch len(args) {
			case 2:
				targetPath = args[0]
				ttlStr = args[1]
				autoName = false
			case 1:
				if _, err := ttl.Parse(args[0]); err == nil {
					ttlStr = args[0]
				} else {
					targetPath = args[0]
					autoName = false
				}
			}

			dur, err := ttl.Parse(ttlStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitUserError)
			}

			err = withRegistry(false, true, func(reg *registry.Registry) error {
				if autoName {
					cwd, err := os.Getwd()
					if err != nil {
						fmt.Fprintf(os.Stderr, "mehdir: getting working directory: %v\n", err)
						os.Exit(exitInternalError)
					}
					targetPath, err = os.MkdirTemp(cwd, reg.GetPrefix())
					if err != nil {
						fmt.Fprintf(os.Stderr, "mehdir: creating temp dir: %v\n", err)
						os.Exit(exitInternalError)
					}
				} else {
					targetPath, err = filepath.Abs(targetPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "mehdir: resolving path: %v\n", err)
						os.Exit(exitInternalError)
					}
					if err := os.MkdirAll(targetPath, 0700); err != nil {
						fmt.Fprintf(os.Stderr, "mehdir: creating directory: %v\n", err)
						os.Exit(exitInternalError)
					}
				}

				if err := os.Chmod(targetPath, 0700); err != nil {
					fmt.Fprintf(os.Stderr, "mehdir: chmod: %v\n", err)
					os.Exit(exitInternalError)
				}

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

			fmt.Println(targetPath)
			return nil
		},
	}
}
