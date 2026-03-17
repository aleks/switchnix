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
	return fmt.Sprintf("ssh -o BatchMode=yes -p %d", h.Port)
}

// RsyncPullArgs returns rsync arguments for pulling from remote to local.
func RsyncPullArgs(h *config.Host, remotePath, localPath string) []string {
	return []string{
		"-avz",
		"--rsync-path=sudo rsync",
		"-e", sshCommandString(h),
		fmt.Sprintf("%s@%s:%s", h.Username, h.Hostname, remotePath),
		localPath,
	}
}

// RsyncPushArgs returns rsync arguments for pushing from local to remote.
func RsyncPushArgs(h *config.Host, localPath, remotePath string) []string {
	return []string{
		"-avz",
		"--delete",
		"--rsync-path=sudo rsync",
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
