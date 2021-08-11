package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"capact.io/capact/cmd/cli/cmd"
)

func main() {
	rootCmd := cmd.NewRoot()
	ctx, cancel := cancelableContext()
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

// cancelableContext returns context that is canceled when stop signal is received
func cancelableContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
		case <-sigCh:
			cancel()
		}
	}()

	return ctx, cancel
}
