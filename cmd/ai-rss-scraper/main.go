// ai-rss-scraper: a program to scrape rss feeds and score them using AI, then email the results
// Scott Baker, https://medium.com/@smbaker
// https://github.com/scottmbaker/ai-rss-scraper

package main

import (
	"github.com/scottmbaker/ai-rss-scraper/internal/commands"
)

func main() {
	commands.Execute()
}
