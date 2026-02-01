package commands

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/scottmbaker/ai-rss-scraper/pkg/report"
	"github.com/scottmbaker/ai-rss-scraper/pkg/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	reportAge       int
	reportThreshold int
	reportOut       string
	reportSendEmail bool
	reportAlways    bool
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate an HTML report of high-scoring articles",
	Run: func(cmd *cobra.Command, args []string) {
		runReport()
	},
}

func init() {
	reportCmd.Flags().IntVar(&reportAge, "age", 7, "Age of articles in days to include in report")
	reportCmd.Flags().IntVar(&reportThreshold, "threshold", 50, "Score threshold for report")
	reportCmd.Flags().StringVar(&reportOut, "out", "report.html", "Output filename for the report")
	reportCmd.Flags().BoolVar(&reportSendEmail, "send-email", false, "Send report via email")
	reportCmd.Flags().BoolVar(&reportAlways, "always", false, "Include articles that have already been reported")
}

func runReport() {
	if err := generateReport(); err != nil {
		log.Fatalf("Error running report: %v", err)
	}
}

func generateReport() error {
	since := time.Now().Add(time.Duration(-reportAge) * 24 * time.Hour)
	var articles []storage.Article
	var err error

	if reportAlways {
		articles, err = DB.GetArticlesAfter(since)
	} else {
		articles, err = DB.GetUnreportedArticlesAfter(since)
	}

	if err != nil {
		return fmt.Errorf("error fetching articles: %v", err)
	}

	if len(articles) == 0 {
		log.Println("No unreported articles found in the specified timeframe.")
		return nil
	}

	var validArticles []storage.Article
	for _, art := range articles {
		score, err := strconv.Atoi(art.Score)
		if err != nil {
			// If the article has no score, or the score is not an integer, then treat it as 0.
			score = 0
		}
		if score >= reportThreshold {
			validArticles = append(validArticles, art)
		}
	}

	if len(validArticles) == 0 {
		log.Println("No articles met the score threshold. Skipping report.")
		return nil
	}

	title := fmt.Sprintf("AI RSS Report (%d days, score >= %d)", reportAge, reportThreshold)

	// Ensure that we're sending the report somehwere
	if !reportSendEmail && reportOut == "" {
		return fmt.Errorf("error: must specify --out or --send-email")
	}

	rep := report.NewReport(title, validArticles)

	// Write to file
	if reportOut != "" {
		if err := rep.GenerateFile(reportOut); err != nil {
			return fmt.Errorf("error writing report file: %v", err)
		}
	}

	// Send email
	if reportSendEmail {
		to := viper.GetString("email_to")
		from := viper.GetString("email_from")
		smarthost := viper.GetString("email_smarthost")
		identity := viper.GetString("email_identity")
		username := viper.GetString("email_username")
		password := viper.GetString("email_password")

		cfgSubject := viper.GetString("email_subject")
		if cfgSubject != "" {
			rep.Title = cfgSubject
		}

		log.Printf("Sending email to %s via %s...", to, smarthost)
		if err := rep.GenerateEmail(to, from, smarthost, identity, username, password); err != nil {
			return fmt.Errorf("error sending email: %v", err)
		}
		log.Println("Email sent successfully.")
	}

	fmt.Printf("Processed %d articles.\n", len(validArticles))

	// Mark articles as reported
	var guids []string
	for _, art := range validArticles {
		guids = append(guids, art.GUID)
	}
	if err := DB.MarkArticlesReported(guids); err != nil {
		return fmt.Errorf("error marking articles as reported: %v", err)
	}
	log.Printf("Marked %d articles as reported.", len(guids))
	return nil
}
