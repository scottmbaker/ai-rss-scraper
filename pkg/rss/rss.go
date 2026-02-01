package rss

import (
	"github.com/mmcdole/gofeed"
)

// TODO: Not much to see here. Consider removing this layer of abstraction.

// FetchFeed fetches and parses an RSS feed from the given URL.
func FetchFeed(url string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	return fp.ParseURL(url)
}
