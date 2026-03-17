package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/aleks/switchnix/internal/config"
)

// SSHArgs returns SSH arguments for non-interactive (batch mode) commands.
// BatchMode=yes ensures SSH never falls back to password auth.
func SSHArgs(h *config.Host) []string {
	args := []string{
		"-o", "BatchMode=yes",
		"-p", strconv.Itoa(h.Port),
	}
	args = append(args, h.SSHOptions...)
	args = append(args, fmt.Sprintf("%s@%s", h.Username, h.Hostname))
	return args
}

// InteractiveSSHArgs returns SSH arguments for interactive commands.
// Uses -t for PTY allocation (needed for sudo). Does not use BatchMode.
func InteractiveSSHArgs(h *config.Host) []string {
	args := []string{
		"-t",
		"-p", strconv.Itoa(h.Port),
	}
	args = append(args, h.SSHOptions...)
	args = append(args, fmt.Sprintf("%s@%s", h.Username, h.Hostname))
	return args
}

// RunSSH executes a command on the remote host and returns combined output.
func RunSSH(ctx context.Context, h *config.Host, command string) ([]byte, error) {
	args := SSHArgs(h)
	args = append(args, command)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	return cmd.CombinedOutput()
}

// RunSSHInteractive executes a command with PTY allocation and stdin/stdout passthrough.
func RunSSHInteractive(ctx context.Context, h *config.Host, command string) error {
	args := InteractiveSSHArgs(h)
	args = append(args, command)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func sshCommandString(h *config.Host) string {
	cmd := fmt.Sprintf("ssh -o BatchMode=yes -p %d", h.Port)
	for _, opt := range h.SSHOptions {
		cmd += " " + opt
	}
	return cmd
}

// StagePull copies /etc/nixos/ to a user-readable temp dir on the remote host
// via interactive SSH (supporting sudo password prompts). It returns the remote
// staging path and a cleanup function.
func StagePull(ctx context.Context, h *config.Host) (remotePath string, cleanup func(), err error) {
	const stageDir = "/tmp/switchnix-stage"

	// Use interactive SSH so sudo can prompt for a password via the TTY.
	stageCmd := fmt.Sprintf(
		"sudo rm -rf %s && sudo cp -a /etc/nixos/ %s && sudo chown -R %s: %s",
		stageDir, stageDir, h.Username, stageDir,
	)
	if err := RunSSHInteractive(ctx, h, stageCmd); err != nil {
		return "", nil, fmt.Errorf("failed to stage remote files: %w", err)
	}

	cleanup = func() {
		// Best-effort cleanup over non-interactive SSH (no sudo needed).
		_, _ = RunSSH(ctx, h, fmt.Sprintf("rm -rf %s", stageDir))
	}

	return stageDir + "/", cleanup, nil
}

// StagePush copies files from a user-writable temp dir into /etc/nixos/ on the
// remote host via interactive SSH (supporting sudo password prompts). It returns
// the remote staging path for the caller to rsync into, and a function that
// performs the final sudo copy + cleanup.
func StagePush(ctx context.Context, h *config.Host) (remotePath string, apply func() error, err error) {
	const stageDir = "/tmp/switchnix-stage"

	// Create the staging directory (no sudo needed).
	output, err := RunSSH(ctx, h, fmt.Sprintf("rm -rf %s && mkdir -p %s", stageDir, stageDir))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create staging directory: %s", string(output))
	}

	apply = func() error {
		// Use interactive SSH so sudo can prompt for a password.
		applyCmd := fmt.Sprintf(
			"sudo rsync -a --delete %s/ /etc/nixos/ && rm -rf %s",
			stageDir, stageDir,
		)
		return RunSSHInteractive(ctx, h, applyCmd)
	}

	return stageDir + "/", apply, nil
}

// RsyncPullArgs returns rsync arguments for pulling from remote to local.
// No sudo needed — the caller should stage files to a user-readable path first.
func RsyncPullArgs(h *config.Host, remotePath, localPath string) []string {
	return []string{
		"-avz",
		"-e", sshCommandString(h),
		fmt.Sprintf("%s@%s:%s", h.Username, h.Hostname, remotePath),
		localPath,
	}
}

// RsyncPushArgs returns rsync arguments for pushing from local to remote.
// No sudo needed — the caller should push to a staging path, then apply.
func RsyncPushArgs(h *config.Host, localPath, remotePath string) []string {
	return []string{
		"-avz",
		"--delete",
		"-e", sshCommandString(h),
		localPath,
		fmt.Sprintf("%s@%s:%s", h.Username, h.Hostname, remotePath),
	}
}

// Rsync runs rsync with the given arguments.
func Rsync(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
