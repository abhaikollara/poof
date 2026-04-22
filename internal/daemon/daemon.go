package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"abhai.dev/poof/internal/registry"
	"abhai.dev/poof/internal/sweep"
)

const pollInterval = 10 * time.Second

func pidDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "poof")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".config", "poof")
}

func PidPath() string {
	return filepath.Join(pidDir(), "daemon.pid")
}

// Running returns the PID of the running daemon, or 0 if not running.
func Running() int {
	data, err := os.ReadFile(PidPath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	// Check if process is alive.
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		// Process doesn't exist, clean up stale PID file.
		os.Remove(PidPath())
		return 0
	}
	return pid
}

// WritePid writes the current process PID to the PID file.
func WritePid() error {
	if err := os.MkdirAll(pidDir(), 0700); err != nil {
		return err
	}
	return os.WriteFile(PidPath(), []byte(strconv.Itoa(os.Getpid())), 0600)
}

// RemovePid removes the PID file.
func RemovePid() {
	os.Remove(PidPath())
}

// Stop sends SIGTERM to the running daemon.
func Stop() error {
	pid := Running()
	if pid == 0 {
		return errors.New("daemon is not running")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM to %d: %w", pid, err)
	}
	// Wait briefly for it to exit, then clean up PID file.
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if Running() == 0 {
			return nil
		}
	}
	RemovePid()
	return nil
}

// Run is the daemon main loop. It blocks until signaled.
func Run(stop <-chan struct{}) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Run once immediately on start.
	doSweep()

	for {
		select {
		case <-ticker.C:
			doSweep()
		case <-stop:
			return
		}
	}
}

func doSweep() {
	lock, err := registry.Lock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "poof-daemon: lock: %v\n", err)
		return
	}
	defer registry.Unlock(lock)

	reg, err := registry.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "poof-daemon: load: %v\n", err)
		return
	}

	before := len(reg.Entries)
	sweep.Run(reg, time.Now().UTC())
	after := len(reg.Entries)

	if before != after {
		if err := registry.Save(reg); err != nil {
			fmt.Fprintf(os.Stderr, "poof-daemon: save: %v\n", err)
		}
	}
}
