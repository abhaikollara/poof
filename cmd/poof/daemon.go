package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"abhai.dev/poof/internal/daemon"
	"github.com/spf13/cobra"
)

const envDaemonChild = "POOF_DAEMON_CHILD"

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the background cleanup daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(daemonStartCmd())
	cmd.AddCommand(daemonStopCmd())
	cmd.AddCommand(daemonStatusCmd())

	return cmd
}

func daemonStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the background cleanup daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pid := daemon.Running(); pid != 0 {
				fmt.Fprintf(os.Stderr, "poof: daemon already running (pid %d)\n", pid)
				return nil
			}

			// If we're the forked child, run the daemon loop.
			if os.Getenv(envDaemonChild) == "1" {
				return runDaemonChild()
			}

			// Fork: re-exec ourselves with the child env var set.
			return forkDaemon()
		},
	}
}

func daemonStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the background cleanup daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := daemon.Stop(); err != nil {
				fmt.Fprintf(os.Stderr, "poof: %v\n", err)
				os.Exit(exitUserError)
			}
			fmt.Fprintln(os.Stderr, "poof: daemon stopped")
			return nil
		},
	}
}

func daemonStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the daemon is running",
		Run: func(cmd *cobra.Command, args []string) {
			if pid := daemon.Running(); pid != 0 {
				fmt.Fprintf(os.Stderr, "poof: daemon running (pid %d)\n", pid)
			} else {
				fmt.Fprintln(os.Stderr, "poof: daemon not running")
			}
		},
	}
}

func forkDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}

	// Open /dev/null for stdin/stdout/stderr of the child.
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("opening /dev/null: %w", err)
	}
	defer devNull.Close()

	env := append(os.Environ(), envDaemonChild+"=1")
	attr := &os.ProcAttr{
		Env:   env,
		Files: []*os.File{devNull, devNull, devNull},
		Sys: &syscall.SysProcAttr{
			Setsid: true, // Detach from terminal.
		},
	}

	proc, err := os.StartProcess(exe, []string{exe, "daemon", "start"}, attr)
	if err != nil {
		return fmt.Errorf("starting daemon: %w", err)
	}
	proc.Release()

	fmt.Fprintf(os.Stderr, "poof: daemon started (pid %d)\n", proc.Pid)
	return nil
}

func runDaemonChild() error {
	if err := daemon.WritePid(); err != nil {
		return fmt.Errorf("writing pid: %w", err)
	}
	defer daemon.RemovePid()

	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		close(stop)
	}()

	daemon.Run(stop)
	return nil
}

// ensureDaemon starts the daemon if it's not already running.
func ensureDaemon() {
	if daemon.Running() != 0 {
		return
	}
	if err := forkDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "poof: warning: could not start daemon: %v\n", err)
	}
}
