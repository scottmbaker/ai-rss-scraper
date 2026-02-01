package commands

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

// NewAIClient creates an OpenAI client using the configured API key and Base URL.
func NewAIClient() (*openai.Client, error) {
	token := viper.GetString("api_key")
	if token == "" {
		return nil, fmt.Errorf("API_KEY is not set; please set it in the config, environment, or command line")
	}

	config := openai.DefaultConfig(token)
	config.BaseURL = viper.GetString("base_url")
	return openai.NewClientWithConfig(config), nil
}
