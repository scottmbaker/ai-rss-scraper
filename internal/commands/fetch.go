package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/scottmbaker/ai-rss-scraper/pkg/rss"
	"github.com/scottmbaker/ai-rss-scraper/pkg/storage"
	"github.com/scottmbaker/ai-rss-scraper/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MAX_DB_CONTENT_LENGTH is the maximum length of the content to be stored in the database.
// This is to prevent the database from growing too large, in case articles are long.
const MAX_DB_CONTENT_LENGTH = 4096

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch articles from RSS feed and save to DB",
	Run: func(cmd *cobra.Command, args []string) {
		runFetch()
	},
}

func runFetch() {
	if err := fetchAndSave(DB, viper.GetString("feed_url")); err != nil {
		log.Printf("Error fetching feed: %v", err)
	}
}

func fetchAndSave(db *storage.DB, url string) error {
	feed, err := rss.FetchFeed(url)
	if err != nil {
		return err
	}

	p := bluemonday.StrictPolicy()
	fmt.Println("Fetching latest items from ", url)

	received := len(feed.Items)
	existing := 0
	added := 0

	for _, item := range feed.Items {
		effectiveGUID := item.GUID
		if effectiveGUID == "" {
			effectiveGUID = item.Link
		}

		exists, err := db.ArticleExists(effectiveGUID)
		if err != nil {
			log.Printf("Error checking DB for article %s: %v", item.Title, err)
			continue
		}
		if exists {
			existing++
			continue
		}

		content := item.Content
		if content == "" {
			content = item.Description
		}
		content = p.Sanitize(content)
		desc := p.Sanitize(item.Description)

		fmt.Printf("- New: %s\n", item.Title)

		pubDate := time.Now()
		if item.PublishedParsed != nil {
			pubDate = *item.PublishedParsed
		}

		// Save new article with empty score/analysis
		art := storage.Article{
			GUID:          effectiveGUID,
			Title:         item.Title,
			Link:          item.Link,
			Description:   desc,
			Content:       utils.TrimString(content, MAX_DB_CONTENT_LENGTH),
			PublishedDate: pubDate,
			Score:         "",
			Analysis:      "",
			FeedURL:       url,
		}
		if err := db.SaveArticle(art); err != nil {
			log.Printf("Error saving article to DB: %v", err)
		} else {
			added++
		}
	}

	fmt.Printf("Fetch Complete: Received: %d, Existing: %d, Added: %d\n", received, existing, added)

	return nil
}
