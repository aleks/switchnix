package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleks/switchnix/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(1)
	}
}
