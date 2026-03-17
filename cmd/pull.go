package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/diff"
	"github.com/aleks/switchnix/internal/ssh"
	"github.com/aleks/switchnix/internal/ui"
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

	fmt.Println(ui.Info.Render(fmt.Sprintf("Pulling configuration from %s (%s)...", host.Name, host.Hostname)))

	ctx, cancel := withTimeout(cmdCtx, 10*time.Minute)
	defer cancel()

	// Stage remote files to a user-readable temp dir on the remote host.
	remotePath, cleanup, err := ssh.StagePull(ctx, host)
	if err != nil {
		return err
	}
	defer cleanup()

	// Rsync into a local temp directory so we can diff before overwriting.
	tmpDir, err := os.MkdirTemp("", "switchnix-pull-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpPath := tmpDir + string(os.PathSeparator)
	rsyncArgs := ssh.RsyncPullArgs(host, remotePath, tmpPath)
	if err := ssh.Rsync(ctx, rsyncArgs); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Read pulled remote files from temp dir.
	remoteFiles, err := readLocalFiles(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read pulled files: %w", err)
	}

	// Read existing local files (may be empty on first pull).
	var localFiles map[string]string
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		localFiles = map[string]string{}
	} else {
		localFiles, err = readLocalFiles(localDir)
		if err != nil {
			return fmt.Errorf("failed to read local files: %w", err)
		}
	}

	// Diff: remote is the "new" state, local is the "old" state.
	cs := diff.ComputeChangeSet(remoteFiles, localFiles)
	if !cs.HasChanges() {
		fmt.Println(ui.Info.Render("No changes to pull."))
		return nil
	}

	diff.PrintChangeSet(cs, remoteFiles, localFiles, "locally")

	// Confirm
	fmt.Printf("Apply these changes to %s? %s ", localDir, ui.Faint.Render("[y/N]:"))
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println(ui.Warn.Render("Cancelled."))
		return nil
	}

	// Apply: rsync from temp dir to real local dir.
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", localDir, err)
	}

	localPath := localDir + string(os.PathSeparator)
	applyArgs := []string{"-avz", "--delete", tmpPath, localPath}
	if err := ssh.Rsync(ctx, applyArgs); err != nil {
		return fmt.Errorf("failed to apply pulled files: %w", err)
	}

	fmt.Println(ui.Success.Render(fmt.Sprintf("Configuration pulled to %s", localDir)))
	return nil
}
