package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"abhai.dev/mehdir/internal/daemon"
	"github.com/spf13/cobra"
)

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the background cleanup daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(daemonInstallCmd())
	cmd.AddCommand(daemonUninstallCmd())
	cmd.AddCommand(daemonStartCmd())
	cmd.AddCommand(daemonStopCmd())
	cmd.AddCommand(daemonStatusCmd())
	cmd.AddCommand(daemonRunCmd())

	return cmd
}

func daemonInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install and start the daemon as a system service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := daemon.Install(); err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitInternalError)
			}
			fmt.Fprintln(os.Stderr, "mehdir: daemon installed and started")
			return nil
		},
	}
}

func daemonUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop and remove the daemon service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := daemon.Uninstall(); err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitInternalError)
			}
			fmt.Fprintln(os.Stderr, "mehdir: daemon uninstalled")
			return nil
		},
	}
}

func daemonStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the daemon service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := daemon.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitUserError)
			}
			fmt.Fprintln(os.Stderr, "mehdir: daemon started")
			return nil
		},
	}
}

func daemonStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the daemon service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := daemon.Stop(); err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitUserError)
			}
			fmt.Fprintln(os.Stderr, "mehdir: daemon stopped")
			return nil
		},
	}
}

func daemonStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the daemon is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := daemon.Status()
			if err != nil {
				fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
				os.Exit(exitInternalError)
			}
			fmt.Fprintf(os.Stderr, "mehdir: daemon %s\n", status)
			return nil
		},
	}
}

// daemonRunCmd runs the daemon in the foreground. Invoked by launchd/systemd.
func daemonRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "run",
		Short:  "Run the daemon in the foreground (used by the service manager)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			stop := make(chan struct{})
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

			go func() {
				<-sigs
				close(stop)
			}()

			daemon.Run(stop)
			return nil
		},
	}
}

// ensureDaemon installs and/or starts the daemon if it's not running.
func ensureDaemon() {
	if !daemon.Installed() {
		if err := daemon.Install(); err != nil {
			fmt.Fprintf(os.Stderr, "mehdir: warning: could not install daemon: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "mehdir: daemon installed and started")
		}
		return
	}
	// Installed but maybe not running — check and restart if needed.
	status, err := daemon.Status()
	if err != nil {
		return
	}
	if status == "not running" || status == "inactive" {
		if err := daemon.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "mehdir: warning: could not start daemon: %v\n", err)
		}
	}
}
