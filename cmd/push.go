package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/diff"
	"github.com/aleks/switchnix/internal/ssh"
	"github.com/aleks/switchnix/internal/ui"
	"github.com/spf13/cobra"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

var pushDryRun bool

var pushCmd = &cobra.Command{
	Use:   "push <host>",
	Short: "Push local NixOS configuration to a remote host",
	Long:  "Shows a diff of changes and, upon confirmation, pushes configurations/<host>/ to /etc/nixos/ on the remote.",
	Args:  cobra.ExactArgs(1),
	RunE:  runPush,
}

func init() {
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "show diff without pushing changes")
	rootCmd.AddCommand(pushCmd)
}

// pushToStaging handles diff display, confirmation, and rsync to staging.
// Returns an apply function that commits staged files to /etc/nixos/ and cleans up.
// If dryRun is true, shows the diff and returns without staging.
func pushToStaging(ctx context.Context, host *config.Host, isDryRun bool) (apply func() error, err error) {
	localDir := filepath.Join("configurations", host.Name)
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no local configuration found at %s. Run 'switchnix pull %s' first", localDir, host.Name)
	}

	// Read local files
	localFiles, err := readLocalFiles(localDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read local files: %w", err)
	}

	// Read remote files (uses interactive sudo for staging)
	fmt.Println(ui.Info.Render(fmt.Sprintf("Fetching remote configuration from %s (%s)...", host.Name, host.Hostname)))

	fetchCtx, fetchCancel := withTimeout(ctx, 5*time.Minute)
	defer fetchCancel()

	remoteFiles, err := readRemoteFiles(fetchCtx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote files: %w", err)
	}

	// Compute and display changes
	cs := diff.ComputeChangeSet(localFiles, remoteFiles)
	if !cs.HasChanges() {
		fmt.Println(ui.Info.Render("No changes to push."))
		return nil, nil
	}

	diff.PrintChangeSet(cs, localFiles, remoteFiles, "on remote")

	if isDryRun {
		fmt.Println(ui.Faint.Render("Dry run — no changes pushed."))
		return nil, nil
	}

	// Confirm
	fmt.Printf("Apply these changes to %s (%s)? %s ", host.Name, host.Hostname, ui.Faint.Render("[y/N]:"))
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println(ui.Warn.Render("Cancelled."))
		return nil, nil
	}

	// Stage push uses the parent context so the apply closure remains valid
	// after this function returns.
	remotePath, applyFn, err := ssh.StagePush(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare remote staging: %w", err)
	}

	// Rsync gets a timeout
	pushCtx, pushCancel := withTimeout(ctx, 10*time.Minute)
	defer pushCancel()

	localPath := localDir + string(os.PathSeparator)
	rsyncArgs := ssh.RsyncPushArgs(host, localPath, remotePath)
	if err := ssh.Rsync(pushCtx, rsyncArgs); err != nil {
		return nil, fmt.Errorf("rsync push failed: %w", err)
	}

	return applyFn, nil
}

func runPush(cmd *cobra.Command, args []string) error {
	hostName := args[0]

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	host, err := cfg.FindHost(hostName)
	if err != nil {
		return err
	}

	apply, err := pushToStaging(cmdCtx, host, pushDryRun)
	if err != nil {
		return err
	}
	if apply == nil {
		return nil // no changes or dry run
	}

	if err := apply(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	fmt.Println(ui.Success.Render(fmt.Sprintf("Configuration pushed to %s successfully.", host.Name)))
	return nil
}

func readLocalFiles(dir string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil // skip symlinks, devices, etc.
		}
		if info.Size() > maxFileSize {
			return fmt.Errorf("file %s exceeds maximum size of %d bytes", path, maxFileSize)
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[relPath] = string(data)
		return nil
	})
	return files, err
}

// safePathRegexp allows only safe relative paths (no shell metacharacters).
// Supports dotfiles (e.g., .gitignore) and paths starting with underscores.
var safePathRegexp = regexp.MustCompile(`^[a-zA-Z0-9._][a-zA-Z0-9._/-]*$`)

func isPathSafe(path string) bool {
	if !safePathRegexp.MatchString(path) {
		return false
	}
	// Reject path traversal
	cleaned := filepath.Clean(path)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return false
	}
	return true
}

func readRemoteFiles(ctx context.Context, host *config.Host) (map[string]string, error) {
	// Stage remote files to a user-readable temp dir (prompts for sudo).
	remotePath, cleanup, err := ssh.StagePull(ctx, host)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// remotePath ends with "/", strip it for find command
	stageDir := strings.TrimSuffix(remotePath, "/")

	// List files (no sudo needed — staged files are user-owned).
	output, err := ssh.RunSSH(ctx, host, fmt.Sprintf("find %s -type f -printf '%%P\\n'", stageDir))
	if err != nil {
		return nil, fmt.Errorf("failed to list remote files: %s", strings.TrimSpace(string(output)))
	}

	files := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !isPathSafe(line) {
			return nil, fmt.Errorf("remote file path %q contains unsafe characters", line)
		}
		content, err := ssh.RunSSH(ctx, host, fmt.Sprintf("cat -- %s/%s", stageDir, line))
		if err != nil {
			return nil, fmt.Errorf("failed to read remote file %s: %s", line, strings.TrimSpace(string(content)))
		}
		files[line] = string(content)
	}

	return files, nil
}
