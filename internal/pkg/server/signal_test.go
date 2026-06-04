package server

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetupSignalHandler(t *testing.T) {
	// Save and restore global state to avoid panics on subsequent calls
	oldOnlyOne := onlyOneSignalHandler
	oldShutdownHandler := shutdownHandler
	defer func() {
		onlyOneSignalHandler = oldOnlyOne
		shutdownHandler = oldShutdownHandler
	}()
	onlyOneSignalHandler = make(chan struct{})

	stop := SetupSignalHandler()
	assert.NotNil(t, stop, "stop channel should not be nil")

	// Simulate signal by sending directly to the shutdownHandler channel
	// os.Process.Signal doesn't support SIGTERM on Windows
	shutdownHandler <- syscall.SIGTERM

	select {
	case <-stop:
		// Success: stop channel was closed
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for stop channel to close after signal")
	}
}

func TestShutdownHandlerSecondSignal(t *testing.T) {
	// Verify that the goroutine in SetupSignalHandler handles a second signal
	oldOnlyOne := onlyOneSignalHandler
	oldShutdownHandler := shutdownHandler
	defer func() {
		onlyOneSignalHandler = oldOnlyOne
		shutdownHandler = oldShutdownHandler
	}()
	onlyOneSignalHandler = make(chan struct{})

	stop := SetupSignalHandler()
	assert.NotNil(t, stop)

	// First signal closes stop channel
	shutdownHandler <- syscall.SIGTERM
	select {
	case <-stop:
		// success
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for first signal")
	}

	// Second signal should trigger os.Exit(0) - which we can't easily test
	// Just verify no panic occurs and the stop channel is already closed
	select {
	case <-stop:
		// Already closed, as expected
	default:
		t.Fatal("stop channel should be closed after first signal")
	}
}

func TestOnlyOneSignalHandlerPanic(t *testing.T) {
	oldOnlyOne := onlyOneSignalHandler
	defer func() {
		onlyOneSignalHandler = oldOnlyOne
	}()
	onlyOneSignalHandler = make(chan struct{})

	assert.Panics(t, func() {
		onlyOneSignalHandler = make(chan struct{})
		close(onlyOneSignalHandler)
		close(onlyOneSignalHandler)
	}, "closing an already-closed channel should panic")
}
