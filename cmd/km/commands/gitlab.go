package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var gitlabCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "Perform keymaster authentication, optimised for gitlab CI.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gitlab command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(gitlabCmd)
}
