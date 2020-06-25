package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage keymaster configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
