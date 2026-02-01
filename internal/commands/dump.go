package commands

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump full details of articles from the database",
	Run: func(cmd *cobra.Command, args []string) {
		runDump()
	},
}

func runDump() {
	// List up to 1000 recent articles
	articles, err := DB.ListArticles(1000)
	if err != nil {
		log.Fatalf("Error listing articles: %v", err)
	}

	for _, art := range articles {
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("Title:       %s\n", art.Title)
		fmt.Printf("GUID:        %s\n", art.GUID)
		fmt.Printf("Date:        %s\n", art.PublishedDate.Format("2006-01-02 15:04:05"))
		fmt.Printf("Link:        %s\n", art.Link)
		fmt.Printf("Feed URL:    %s\n", art.FeedURL)
		fmt.Printf("Score:       %s\n", art.Score)
		fmt.Printf("Model:       %s\n", art.Model)
		fmt.Println("Analysis:")
		fmt.Println(art.Analysis)
		fmt.Println("Description:")
		fmt.Println(art.Description)
		fmt.Println("Content:")
		fmt.Println(art.Content)
		fmt.Println()
	}
}
