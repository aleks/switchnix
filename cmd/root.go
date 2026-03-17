package cmd

import (
	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "switchnix",
	Short: "Manage NixOS configurations on remote servers",
	Long:  "A CLI tool to pull, push, and apply NixOS configurations on remote servers via SSH.",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "hosts.yml", "path to hosts configuration file")
}

func Execute() error {
	return rootCmd.Execute()
}
