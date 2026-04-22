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
If no arguments are given, an auto-named directory is created with the default TTL.

The auto-generated prefix and default TTL can be changed with "mehdir config".`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetPath string
			var ttlStr string
			autoName := true
			ttlExplicit := false

			switch len(args) {
			case 2:
				targetPath = args[0]
				ttlStr = args[1]
				autoName = false
				ttlExplicit = true
			case 1:
				// If it exists on disk, treat as a path even if it parses as a TTL.
				// e.g. "2d" could be a directory name or "2 days".
				if _, statErr := os.Stat(args[0]); statErr == nil {
					targetPath = args[0]
					autoName = false
				} else if _, err := ttl.Parse(args[0]); err == nil {
					ttlStr = args[0]
					ttlExplicit = true
				} else {
					targetPath = args[0]
					autoName = false
				}
			}

			var createdPath string

			err := withRegistry(false, true, func(reg *registry.Registry) error {
				// Resolve TTL: use explicit value, or fall back to config default.
				if !ttlExplicit {
					ttlStr = reg.GetTTL()
				}

				dur, err := ttl.Parse(ttlStr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
					os.Exit(exitUserError)
				}

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
				createdPath = targetPath
				return nil
			})
			if err != nil {
				return err
			}

			ensureDaemon()

			fmt.Println(createdPath)
			return nil
		},
	}
}
