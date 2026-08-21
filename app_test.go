package main

import (
	"context"
	"testing"
)

// The bound methods all take their context from a.context(). Wails calls
// startup before the frontend can reach any of them, but a method reached
// before that must not hand a nil context to os/exec, so the fallback here is
// the one branch in this file worth pinning down.
func TestContextFallsBackWhenStartupHasNotRun(t *testing.T) {
	app := NewApp()

	if got := app.context(); got == nil {
		t.Fatal("context() returned nil before startup")
	}
}

func TestContextReturnsStartupContext(t *testing.T) {
	app := NewApp()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.startup(ctx)

	if got := app.context(); got != ctx {
		t.Errorf("context() = %v, want the context passed to startup", got)
	}
}
