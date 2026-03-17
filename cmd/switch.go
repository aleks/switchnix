package cmd

import (
	"fmt"
	"strings"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/ssh"
	"github.com/aleks/switchnix/internal/ui"
	"github.com/spf13/cobra"
)

var (
	switchAction   string
	switchNoPush   bool
	switchDryRun   bool
	switchNixosArgs []string
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
	switchCmd.Flags().StringSliceVar(&switchNixosArgs, "nixos-args", nil, "additional flags to pass to nixos-rebuild (e.g. --nixos-args='--flake,/etc/nixos#host')")
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

	extraArgs := ""
	if len(switchNixosArgs) > 0 {
		extraArgs = " " + strings.Join(switchNixosArgs, " ")
	}

	if switchNoPush {
		// --no-push: just run rebuild against /etc/nixos/ (original behavior)
		rebuildCmd := fmt.Sprintf("sudo nixos-rebuild %s%s", switchAction, extraArgs)
		fmt.Println(ui.Info.Render(fmt.Sprintf("Running '%s' on %s (%s)...", rebuildCmd, host.Name, host.Hostname)))
		fmt.Println()

		if err := ssh.RunSSHInteractive(cmdCtx, host, rebuildCmd); err != nil {
			return fmt.Errorf("nixos-rebuild %s failed: %w", switchAction, err)
		}

		fmt.Println()
		fmt.Println(ui.Success.Render(fmt.Sprintf("nixos-rebuild %s completed successfully on %s.", switchAction, host.Name)))
		return nil
	}

	// Atomic push+switch: push to staging, rebuild from staging, commit on success
	apply, err := pushToStaging(cmdCtx, host, switchDryRun)
	if err != nil {
		return err
	}
	if apply == nil && switchDryRun {
		return nil
	}

	if apply != nil {
		// We have changes — rebuild from staging, commit on success
		defer func() {
			_, _ = ssh.RunSSH(cmdCtx, host, fmt.Sprintf("rm -rf %s", stagingDir))
		}()

		rebuildCmd := fmt.Sprintf("sudo nixos-rebuild %s -I nixos-config=%s/configuration.nix%s", switchAction, stagingDir, extraArgs)
		fmt.Println()
		fmt.Println(ui.Info.Render(fmt.Sprintf("Running '%s' on %s (%s)...", rebuildCmd, host.Name, host.Hostname)))
		fmt.Println()

		if err := ssh.RunSSHInteractive(cmdCtx, host, rebuildCmd); err != nil {
			return fmt.Errorf("nixos-rebuild %s failed (remote config unchanged): %w", switchAction, err)
		}

		if err := apply(); err != nil {
			return fmt.Errorf("rebuild succeeded but failed to commit config: %w", err)
		}
	} else {
		// No changes to push — rebuild from current remote configuration
		rebuildCmd := fmt.Sprintf("sudo nixos-rebuild %s%s", switchAction, extraArgs)
		fmt.Println()
		fmt.Println(ui.Info.Render(fmt.Sprintf("Running '%s' on %s (%s)...", rebuildCmd, host.Name, host.Hostname)))
		fmt.Println()

		if err := ssh.RunSSHInteractive(cmdCtx, host, rebuildCmd); err != nil {
			return fmt.Errorf("nixos-rebuild %s failed: %w", switchAction, err)
		}
	}

	fmt.Println()
	fmt.Println(ui.Success.Render(fmt.Sprintf("nixos-rebuild %s completed successfully on %s.", switchAction, host.Name)))
	return nil
}
