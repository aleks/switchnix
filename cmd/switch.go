package cmd

import (
	"fmt"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/ssh"
	"github.com/spf13/cobra"
)

var switchAction string

var switchCmd = &cobra.Command{
	Use:   "switch <host>",
	Short: "Apply NixOS configuration on a remote host",
	Long:  "Connects via SSH and runs nixos-rebuild on the remote host. Streams output and supports interactive sudo.",
	Args:  cobra.ExactArgs(1),
	RunE:  runSwitch,
}

func init() {
	switchCmd.Flags().StringVar(&switchAction, "action", "switch", "nixos-rebuild action: switch, test, or boot")
	rootCmd.AddCommand(switchCmd)
}

var validActions = map[string]bool{
	"switch": true,
	"test":   true,
	"boot":   true,
}

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

	rebuildCmd := fmt.Sprintf("sudo nixos-rebuild %s", switchAction)
	fmt.Printf("Running '%s' on %s (%s)...\n\n", rebuildCmd, host.Name, host.Hostname)

	// No timeout for switch — nixos-rebuild can take a very long time.
	// Signal handling (Ctrl+C) is still active via cmdCtx.
	if err := ssh.RunSSHInteractive(cmdCtx, host, rebuildCmd); err != nil {
		return fmt.Errorf("nixos-rebuild %s failed: %w", switchAction, err)
	}

	fmt.Printf("\nnixos-rebuild %s completed successfully on %s.\n", switchAction, host.Name)
	return nil
}
