package main

import (
	"fmt"
	"os"
	"time"

	"abhai.dev/mehdir/internal/registry"
	"abhai.dev/mehdir/internal/sweep"
)

const (
	exitUserError     = 1
	exitInternalError = 2
)

// withRegistry acquires the lock, loads the registry, calls fn, saves, and unlocks.
// If sweepBefore is true, expired entries are swept before calling fn.
// If sweepAfter is true, expired entries are swept after calling fn.
func withRegistry(sweepBefore, sweepAfter bool, fn func(reg *registry.Registry) error) error {
	lock, err := registry.Lock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
		os.Exit(exitInternalError)
	}
	defer registry.Unlock(lock)

	reg, err := registry.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
		os.Exit(exitInternalError)
	}

	if sweepBefore {
		sweep.Run(reg, time.Now().UTC())
	}

	if err := fn(reg); err != nil {
		return err
	}

	if sweepAfter {
		sweep.Run(reg, time.Now().UTC())
	}

	if err := registry.Save(reg); err != nil {
		fmt.Fprintf(os.Stderr, "mehdir: %v\n", err)
		os.Exit(exitInternalError)
	}
	return nil
}
