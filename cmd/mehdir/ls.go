package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"abhai.dev/mehdir/internal/registry"
	"github.com/spf13/cobra"
)

func lsCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List active temporary directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRegistry(true, false, func(reg *registry.Registry) error {
				if len(reg.Entries) == 0 {
					fmt.Fprintln(os.Stderr, "mehdir: no active temp dirs")
					return nil
				}

				if jsonOut {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(reg.Entries)
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "PATH\tCREATED\tEXPIRES IN")
				now := time.Now().UTC()
				for _, e := range reg.Entries {
					remaining := e.ExpiresAt.Sub(now)
					fmt.Fprintf(w, "%s\t%s\t%s\n",
						e.Path,
						e.CreatedAt.Local().Format("2006-01-02 15:04"),
						formatDuration(remaining),
					)
				}
				return w.Flush()
			})
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "expired"
	}
	totalSeconds := int(math.Round(d.Seconds()))
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
