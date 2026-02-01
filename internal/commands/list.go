package commands

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent articles and their scores",
	Run: func(cmd *cobra.Command, args []string) {
		runList()
	},
}

func runList() {
	articles, err := DB.ListArticles(1000) // TODO: make this configurable
	if err != nil {
		log.Fatalf("Error listing articles: %v", err)
	}

	for _, art := range articles {
		score := art.Score
		if score == "" {
			score = "---"
		}
		fmt.Printf("[%3s] %s (%s)\n", score, art.Title, art.PublishedDate.Format("2006-01-02"))
	}
}
