package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/scottmbaker/ai-rss-scraper/pkg/storage"
	"github.com/scottmbaker/ai-rss-scraper/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile must remain StringVar because it's used before Viper is fully loaded
	cfgFile string
	DB      *storage.DB
	rootCmd = &cobra.Command{
		Use:   "ai-rss-scraper",
		Short: "A CLI tool to scrape RSS feeds and analyze them with AI",
		Long: `ai-rss-scraper reads an RSS feed,
sanitizes the content, and uses an LLM to score the articles based on
specific preferences. It stores the history in a local SQLite database.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			expectedDBPath := viper.GetString("db_path")
			// Initialize/open the database prior to every command.
			DB, err = storage.NewDatabase(expectedDBPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			return nil
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if DB != nil {
		DB.Close()
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-rss-scraper.yaml)")
	rootCmd.PersistentFlags().String("db-path", "rss_history.db", "SQLite database file path")
	rootCmd.PersistentFlags().String("api-key", "", "POE API Key")
	rootCmd.PersistentFlags().String("base-url", "https://api.poe.com/v1", "OpenAI Compatiable API URL")
	rootCmd.PersistentFlags().String("model", "gemini-3-flash", "LLM Model to use")
	rootCmd.PersistentFlags().String("feed-url", "https://hackaday.com/blog/feed/", "RSS Feed URL")
	rootCmd.PersistentFlags().String("prompt", "", "AI Prompt (string or @filename)")
	rootCmd.PersistentFlags().String("email-smarthost", "", "SMTP Smarthost (hostname:port)")
	rootCmd.PersistentFlags().String("email-identity", "", "Email Identity (Auth Username)")
	rootCmd.PersistentFlags().String("email-username", "", "Email Username")
	rootCmd.PersistentFlags().String("email-password", "", "Email Password")
	rootCmd.PersistentFlags().String("email-to", "", "Email To Address")
	rootCmd.PersistentFlags().String("email-from", "", "Email From Address")
	rootCmd.PersistentFlags().String("email-subject", "rss article scrape results", "Email Subject")

	// Ckerr to make linter happy... is there any real chance of these failing??

	utils.Ckerr(viper.BindPFlag("db_path", rootCmd.PersistentFlags().Lookup("db-path")))
	utils.Ckerr(viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key")))
	utils.Ckerr(viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url")))
	utils.Ckerr(viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model")))
	utils.Ckerr(viper.BindPFlag("feed_url", rootCmd.PersistentFlags().Lookup("feed-url")))
	utils.Ckerr(viper.BindPFlag("prompt", rootCmd.PersistentFlags().Lookup("prompt")))
	utils.Ckerr(viper.BindPFlag("email_smarthost", rootCmd.PersistentFlags().Lookup("email-smarthost")))
	utils.Ckerr(viper.BindPFlag("email_identity", rootCmd.PersistentFlags().Lookup("email-identity")))
	utils.Ckerr(viper.BindPFlag("email_username", rootCmd.PersistentFlags().Lookup("email-username")))
	utils.Ckerr(viper.BindPFlag("email_password", rootCmd.PersistentFlags().Lookup("email-password")))
	utils.Ckerr(viper.BindPFlag("email_to", rootCmd.PersistentFlags().Lookup("email-to")))
	utils.Ckerr(viper.BindPFlag("email_from", rootCmd.PersistentFlags().Lookup("email-from")))
	utils.Ckerr(viper.BindPFlag("email_subject", rootCmd.PersistentFlags().Lookup("email-subject")))

	utils.Ckerr(viper.BindEnv("db_path", "DB_PATH"))
	utils.Ckerr(viper.BindEnv("api_key", "API_KEY"))
	utils.Ckerr(viper.BindEnv("base_url", "BASE_URL"))
	utils.Ckerr(viper.BindEnv("model", "MODEL"))
	utils.Ckerr(viper.BindEnv("feed_url", "FEED_URL"))
	utils.Ckerr(viper.BindEnv("prompt", "PROMPT"))
	utils.Ckerr(viper.BindEnv("email_smarthost", "EMAIL_SMARTHOST"))
	utils.Ckerr(viper.BindEnv("email_identity", "EMAIL_IDENTITY"))
	utils.Ckerr(viper.BindEnv("email_username", "EMAIL_USERNAME"))
	utils.Ckerr(viper.BindEnv("email_password", "EMAIL_PASSWORD"))
	utils.Ckerr(viper.BindEnv("email_to", "EMAIL_TO"))
	utils.Ckerr(viper.BindEnv("email_from", "EMAIL_FROM"))
	utils.Ckerr(viper.BindEnv("email_subject", "EMAIL_SUBJECT"))

	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(scoreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(dumpCmd)
	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(resetReportedCmd)
	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	usingDefaultConfig := false
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ai-rss-scraper")

		usingDefaultConfig = true
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok && usingDefaultConfig {
		// No default config file found, that's fine.
	} else {
		log.Fatalf("Error reading config file: %s\n", err)
	}
}
