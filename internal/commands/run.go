package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/scottmbaker/ai-rss-scraper/internal/htmlserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch articles and then score them",
	Run: func(cmd *cobra.Command, args []string) {
		runScraper()
	},
}

var runInterval time.Duration
var noScore bool
var noFetch bool
var noReport bool

var runServe bool

func init() {
	runCmd.Flags().IntVar(&reportAge, "age", 7, "Age of articles in days to include in report")
	runCmd.Flags().IntVar(&reportThreshold, "threshold", 50, "Score threshold for report")
	runCmd.Flags().StringVar(&reportOut, "out", "", "Output filename for the report")
	runCmd.Flags().BoolVar(&reportSendEmail, "send-email", false, "Send report via email")
	runCmd.Flags().DurationVar(&runInterval, "interval", 0, "Interval to run the scraper loop (e.g. 1h, 30m). 0 means run once.")
	runCmd.Flags().BoolVar(&noFetch, "no-fetch", false, "Don't fetch new articles")
	runCmd.Flags().BoolVar(&noScore, "no-score", false, "Don't score articles")
	runCmd.Flags().BoolVar(&noReport, "no-report", false, "Don't generate report")

	runCmd.Flags().BoolVar(&runServe, "serve", false, "Run the web server")
	runCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "Host interface to listen on")
	runCmd.Flags().IntVar(&servePort, "port", 8080, "Port to listen on")
}

// runScraper is fetch + score + report, and can be set to run in a loop, forever.
// Generally when something goes bad, we do a log.Fatalf() and bail. If we're running
// in Kubernetes, then Kubernetes will restart us.
func runScraper() {
	if runServe {
		server := htmlserver.NewServer(serveHost, servePort, DB)
		go func() {
			if err := server.Start(); err != nil {
				log.Fatalf("Error starting server: %v", err)
			}
		}()
		// Give the server a moment to start
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Starting ai-rss-scraper...")

	for {
		log.Println("Starting cycle...") // Added: Log for cycle start
		if !noFetch {
			if err := fetchAndSave(DB, viper.GetString("feed_url")); err != nil {
				log.Fatalf("Error fetching feed: %v", err)
			}
		}

		if !noScore {
			if err := scoreArticles(DB, "", false); err != nil {
				log.Fatalf("Error scoring articles: %v", err)
			}
		}

		if !noReport {
			if err := generateReport(); err != nil {
				log.Fatalf("Error running report: %v", err)
			}
		}

		if runInterval == 0 {
			// If --serve has been used, but runInterval is 0, then block this loop forever
			// and keep serving in the background.
			if runServe {
				select {}
			}
			// Here is where we exit if runInterval == 0
			break
		}

		log.Printf("Sleeping for %v...", runInterval)
		time.Sleep(runInterval)
	}

	log.Println("ai-rss-scraper finished")
}
