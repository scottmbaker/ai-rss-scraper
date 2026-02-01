package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var resetReportedCmd = &cobra.Command{
	Use:   "reset-reported [pattern]",
	Short: "Reset reported flag for articles matching a pattern",
	Long:  `Reset reported flag for articles matching a pattern. Use "*" as a wildcard.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		affected, err := DB.ResetReported(pattern)
		if err != nil {
			log.Fatalf("Error resetting reported flags: %v", err)
		}
		log.Printf("Reset reported flag for %d articles matching '%s'", affected, pattern)
	},
}
