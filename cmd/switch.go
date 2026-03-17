package cmd

import (
	"fmt"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/ssh"
	"github.com/spf13/cobra"
)

var (
	switchAction string
	switchNoPush bool
	switchDryRun bool
)

var switchCmd = &cobra.Command{
	Use:   "switch <host>",
	Short: "Push configuration and apply it on a remote host",
	Long: `Atomically pushes local configuration to a staging directory on the remote host,
runs nixos-rebuild against the staged config, and only commits to /etc/nixos/ if
the rebuild succeeds. Use --no-push to skip pushing and rebuild from the existing
remote configuration.`,
	Args: cobra.ExactArgs(1),
	RunE: runSwitch,
}

func init() {
	switchCmd.Flags().StringVar(&switchAction, "action", "switch", "nixos-rebuild action: switch, test, or boot")
	switchCmd.Flags().BoolVar(&switchNoPush, "no-push", false, "skip pushing; rebuild from current remote configuration")
	switchCmd.Flags().BoolVar(&switchDryRun, "dry-run", false, "show diff without pushing or switching")
	rootCmd.AddCommand(switchCmd)
}

var validActions = map[string]bool{
	"switch": true,
	"test":   true,
	"boot":   true,
}

const stagingDir = "/tmp/switchnix-stage"

func runSwitch(cmd *cobra.Command, args []string) error {
	hostName := args[0]

	if !validActions[switchAction] {
		return fmt.Errorf("invalid action %q: must be one of: switch, test, boot", switchAction)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	host, err := cfg.FindHost(hostName)
	if err != nil {
		return err
	}

	if switchNoPush {
		// --no-push: just run rebuild against /etc/nixos/ (original behavior)
		rebuildCmd := fmt.Sprintf("sudo nixos-rebuild %s", switchAction)
		fmt.Printf("Running '%s' on %s (%s)...\n\n", rebuildCmd, host.Name, host.Hostname)

		if err := ssh.RunSSHInteractive(cmdCtx, host, rebuildCmd); err != nil {
			return fmt.Errorf("nixos-rebuild %s failed: %w", switchAction, err)
		}

		fmt.Printf("\nnixos-rebuild %s completed successfully on %s.\n", switchAction, host.Name)
		return nil
	}

	// Atomic push+switch: push to staging, rebuild from staging, commit on success
	apply, err := pushToStaging(cmdCtx, host, switchDryRun)
	if err != nil {
		return err
	}
	if apply == nil {
		return nil // no changes or dry run
	}

	// Ensure staging is cleaned up on failure
	defer func() {
		_, _ = ssh.RunSSH(cmdCtx, host, fmt.Sprintf("rm -rf %s", stagingDir))
	}()

	// Build from staging
	rebuildCmd := fmt.Sprintf("sudo nixos-rebuild %s -I nixos-config=%s/configuration.nix", switchAction, stagingDir)
	fmt.Printf("\nRunning '%s' on %s (%s)...\n\n", rebuildCmd, host.Name, host.Hostname)

	if err := ssh.RunSSHInteractive(cmdCtx, host, rebuildCmd); err != nil {
		return fmt.Errorf("nixos-rebuild %s failed (remote config unchanged): %w", switchAction, err)
	}

	// Success — commit staging to /etc/nixos/
	if err := apply(); err != nil {
		return fmt.Errorf("rebuild succeeded but failed to commit config: %w", err)
	}

	fmt.Printf("\nnixos-rebuild %s completed successfully on %s.\n", switchAction, host.Name)
	return nil
}
