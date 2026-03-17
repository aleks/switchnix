package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/ssh"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <host>",
	Short: "Pull NixOS configuration from a remote host",
	Long:  "Connects via SSH and pulls /etc/nixos/ contents into configurations/<host>/.",
	Args:  cobra.ExactArgs(1),
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	hostName := args[0]

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	host, err := cfg.FindHost(hostName)
	if err != nil {
		return err
	}

	localDir := filepath.Join("configurations", host.Name)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", localDir, err)
	}

	// Ensure local path ends with / for rsync
	localPath := localDir + string(os.PathSeparator)

	fmt.Printf("Pulling configuration from %s (%s)...\n", host.Name, host.Hostname)

	ctx, cancel := withTimeout(cmdCtx, 10*time.Minute)
	defer cancel()

	rsyncArgs := ssh.RsyncPullArgs(host, "/etc/nixos/", localPath)
	if err := ssh.Rsync(ctx, rsyncArgs); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Printf("Configuration pulled to %s\n", localDir)
	return nil
}
