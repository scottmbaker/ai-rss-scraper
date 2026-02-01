package commands

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	openai "github.com/sashabaranov/go-openai"
	"github.com/scottmbaker/ai-rss-scraper/pkg/storage"
	"github.com/scottmbaker/ai-rss-scraper/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	refreshPattern string
	showResponse   bool
)

const MAX_AI_CONTENT_LENGTH = 4096

// defaultPromptTemplate is the default prompt, designed for my needs.
// You can override it with the --prompt flag or related environment variable.
const defaultPromptTemplate = "Scott likes projects relating to vintage computers and speech synthesizers. " +
	"In particular he like the 4004, 8008, 8080, 8085, 8086, z80, and z8000 cpus. " +
	"He ikes unique display technologies like nixie tubes. " +
	"He likes raspberry pi and microcontroller projects if there is something unique or retro about them. " +
	"He likes restoring old or rare computers. Produce a numeric score between 0 and 100 " +
	"If the article is about Scott Baker or smbaker, give it a score of 100. " +
	"based on how much scott will like this project, please exactly three bullet points on what he will like.\n\n" +
	"Title: {{.Title}}\nDescription: {{.Description}}\nContent: {{.Content}}"

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "Score unscored articles in DB",
	Run: func(cmd *cobra.Command, args []string) {
		runScore()
	},
}

func init() {
	scoreCmd.Flags().StringVar(&refreshPattern, "refresh", "", "Refresh/Score articles matching title wildcard (e.g. '*Retro*')")
	scoreCmd.Flags().BoolVar(&showResponse, "showresponse", false, "Show the raw response from the model")
}

func runScore() {
	if err := scoreArticles(DB, refreshPattern, showResponse); err != nil {
		log.Fatalf("error scoring articles: %v", err)
	}
}

func scoreArticles(db *storage.DB, refreshPattern string, showResponse bool) error {
	articles, err := db.GetArticlesToScore(refreshPattern)
	if err != nil {
		return err
	}

	if len(articles) == 0 {
		fmt.Println("No unscored articles found.")
		return nil
	}

	client, err := NewAIClient()
	if err != nil {
		return err
	}
	model := viper.GetString("model")

	fmt.Printf("found %d unscored articles\n", len(articles))

	promptTemplate := viper.GetString("prompt")
	if promptTemplate == "" {
		promptTemplate = defaultPromptTemplate
	} else if strings.HasPrefix(promptTemplate, "@") {
		filename := strings.TrimPrefix(promptTemplate, "@")
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading prompt file %s: %w", filename, err)
		}
		promptTemplate = string(content)
	}

	// Parse the template once
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return fmt.Errorf("error parsing prompt template: %w", err)
	}

	scoreRegex := regexp.MustCompile(`(?i)(?:score|rating):\s*(\d+)`)

	for _, art := range articles {
		fmt.Printf("Scoring: %s\n", art.Title)

		data := struct {
			Title       string
			Description string
			Content     string
		}{
			Title:       art.Title,
			Description: art.Description,
			Content:     utils.TrimString(art.Content, MAX_AI_CONTENT_LENGTH),
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			log.Printf("  Error executing prompt template: %v", err)
			continue
		}
		articlePrompt := buf.String()

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: articlePrompt,
					},
				},
			},
		)

		if err != nil {
			log.Printf("  Error calling AI: %v", err)
			continue
		}

		content := resp.Choices[0].Message.Content
		if showResponse {
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Println(content)
			fmt.Println("--------------------------------------------------------------------------------")
		}

		// Attempt to extract score
		scoreMatch := scoreRegex.FindStringSubmatch(content)
		score := "N/A"
		if len(scoreMatch) > 1 {
			score = scoreMatch[1]
		} else {
			// Fallback: find the first number in the text
			fallbackRegex := regexp.MustCompile(`(\d+)`)
			fallbackMatch := fallbackRegex.FindStringSubmatch(content)
			if len(fallbackMatch) > 1 {
				score = fallbackMatch[1]
			}
		}

		fmt.Printf("  Score: %s\n", score)

		if err := db.UpdateArticleScore(art.GUID, score, content, model); err != nil {
			log.Printf("Error updating score for %s: %v", art.Title, err)
		}
	}
	return nil
}
