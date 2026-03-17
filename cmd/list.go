package cmd

import (
	"fmt"

	"github.com/aleks/switchnix/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured hosts",
	Long:  "Displays all hosts defined in hosts.yml.",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	if len(cfg.Hosts) == 0 {
		fmt.Println("No hosts configured. Edit hosts.yml to add hosts.")
		return nil
	}

	fmt.Printf("%-20s %-30s %-15s %s\n", "NAME", "HOSTNAME", "USERNAME", "PORT")
	for _, h := range cfg.Hosts {
		fmt.Printf("%-20s %-30s %-15s %d\n", h.Name, h.Hostname, h.Username, h.Port)
	}
	return nil
}
