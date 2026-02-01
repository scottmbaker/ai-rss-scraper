package commands

import (
	"log"

	"github.com/scottmbaker/ai-rss-scraper/internal/htmlserver"
	"github.com/spf13/cobra"
)

var (
	serveHost string
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server interface",
	Run: func(cmd *cobra.Command, args []string) {
		server := htmlserver.NewServer(serveHost, servePort, DB)
		if err := server.Start(); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "Host interface to listen on")
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to listen on")
}
