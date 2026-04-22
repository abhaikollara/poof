package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"abhai.dev/mehdir/internal/registry"
	"github.com/spf13/cobra"
)

func cleanslateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanslate",
		Short: "Delete all tracked directories and clear the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(false, false, func(reg *registry.Registry) error {
				if len(reg.Entries) == 0 {
					fmt.Fprintln(os.Stderr, "mehdir: nothing to clean up")
					return nil
				}

				fmt.Fprintf(os.Stderr, "This will delete %d tracked directories:\n", len(reg.Entries))
				for _, e := range reg.Entries {
					fmt.Fprintf(os.Stderr, "  %s\n", e.Path)
				}
				fmt.Fprint(os.Stderr, "\nAre you sure? [y/N] ")

				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))

				if answer != "y" && answer != "yes" {
					fmt.Fprintln(os.Stderr, "mehdir: aborted")
					return nil
				}

				removed := 0
				var kept []registry.Entry
				for _, e := range reg.Entries {
					if err := registry.SafeToDelete(e.Path, reg.AllowedPrefixes); err != nil {
						fmt.Fprintf(os.Stderr, "mehdir: skipping %s: %v\n", e.Path, err)
						kept = append(kept, e)
						continue
					}
					if err := os.RemoveAll(e.Path); err != nil {
						fmt.Fprintf(os.Stderr, "mehdir: removing %s: %v\n", e.Path, err)
						kept = append(kept, e)
						continue
					}
					removed++
				}

				reg.Entries = kept
				fmt.Fprintf(os.Stderr, "mehdir: removed %d directories\n", removed)
				if len(kept) > 0 {
					fmt.Fprintf(os.Stderr, "mehdir: %d entries could not be removed\n", len(kept))
				}
				return nil
			})
		},
	}
}
