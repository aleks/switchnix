package cmd

import (
	"fmt"

	"github.com/aleks/switchnix/internal/config"
	"github.com/aleks/switchnix/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
		fmt.Println(ui.Warn.Render("No hosts configured. Edit hosts.yml to add hosts."))
		return nil
	}

	rows := make([][]string, len(cfg.Hosts))
	for i, h := range cfg.Hosts {
		rows[i] = []string{h.Name, h.Hostname, h.Username, fmt.Sprintf("%d", h.Port)}
	}

	t := table.New().
		Headers("NAME", "HOSTNAME", "USERNAME", "PORT").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return ui.Bold
			}
			return lipgloss.NewStyle()
		})

	fmt.Println(t)
	return nil
}
