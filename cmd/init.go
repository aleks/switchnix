package cmd

import (
	"fmt"
	"os"

	"github.com/aleks/switchnix/internal/scaffold"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new switchnix project",
	Long:  "Creates a hosts.yml configuration file and a configurations/ directory in the current directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		if err := scaffold.Init(dir); err != nil {
			return err
		}

		fmt.Println("Initialized switchnix project:")
		fmt.Println("  hosts.yml        - define your remote NixOS hosts here")
		fmt.Println("  configurations/  - host configurations will be stored here")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Add your hosts to hosts.yml")
		fmt.Println("  2. Run 'switchnix pull <host>' to fetch existing configurations")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
