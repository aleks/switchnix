package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

var (
	configPath string
	// Version is set at build time via -ldflags.
	Version = "dev"
)

// cmdCtx holds the context passed from main for signal handling.
var cmdCtx context.Context

var rootCmd = &cobra.Command{
	Use:          "switchnix",
	Short:        "Manage NixOS configurations on remote servers",
	Long:         "A CLI tool to pull, push, and apply NixOS configurations on remote servers via SSH.",
	Version:      Version,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "hosts.yml", "path to hosts configuration file")
}

func Execute(ctx context.Context) error {
	cmdCtx = ctx
	return rootCmd.Execute()
}

// withTimeout creates a child context with a timeout from the parent context.
func withTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, d)
}
