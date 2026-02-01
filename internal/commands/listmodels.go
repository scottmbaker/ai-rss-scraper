package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var listModelsCmd = &cobra.Command{
	Use:   "listmodels",
	Short: "List available models from the AI provider",
	Run: func(cmd *cobra.Command, args []string) {
		runListModels()
	},
}

// runListModels lists the available models from the AI provider.
// This is useful when switching providers, as not all providers name them
// the same way.
func runListModels() {
	client, err := NewAIClient()
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	models, err := client.ListModels(context.Background())
	if err != nil {
		log.Fatalf("Error listing models: %v", err)
	}

	fmt.Printf("Found %d models:\n", len(models.Models))
	for _, m := range models.Models {
		fmt.Printf("- %s (Owner: %s)\n", m.ID, m.OwnedBy)
	}
}
