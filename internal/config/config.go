package config

import (
	"fmt"
	"os"
	"time"
)

type GroupMeConfig struct {
	BotID   string        `json:"bot_id"`
	Token   string        `json:"token"`
	Timeout time.Duration `json:"timeout"`
	Host    string        `json:"host"`
	Path    string        `json:"path"`
}

type OpenAIConfig struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	AssistantID string        `json:"assistant_id"`
	Timeout     time.Duration `json:"timeout"`
}

type Config struct {
	OpenAI  *OpenAIConfig  `json:"openai"`
	GroupMe *GroupMeConfig `json:"groupme"`
	Auth    *AuthConfig    `json:"auth"`
}

type AuthConfig struct {
	APIKey string `json:"api_key"`
}

func LoadConfig() (*Config, error) {
	openAIConfig, err := loadOpenAIConfig()

	if err != nil {
		return nil, err
	}

	groupMeConfig, err := loadGroupMeConfig()

	if err != nil {
		return nil, err
	}

	authConfig, err := loadAuthConfig()

	if err != nil {
		return nil, err
	}

	return &Config{
		OpenAI:  openAIConfig,
		GroupMe: groupMeConfig,
		Auth:    authConfig,
	}, nil
}

func loadOpenAIConfig() (*OpenAIConfig, error) {
	APIKey := os.Getenv("OPENAI_API_KEY")

	if APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	AssistantID := os.Getenv("ASSISTANT_ID")

	if AssistantID == "" {
		return nil, fmt.Errorf("ASSISTANT_ID environment variable is not set")
	}

	return &OpenAIConfig{
		APIKey:      APIKey,
		BaseURL:     "https://api.openai.com/v1",
		AssistantID: AssistantID,
		Timeout:     60 * time.Second,
	}, nil
}

func loadGroupMeConfig() (*GroupMeConfig, error) {
	BotID := os.Getenv("GROUPME_BOT_ID")

	if BotID == "" {
		return nil, fmt.Errorf("GROUPME_BOT_ID environment variable is not set")
	}

	token := os.Getenv("GROUPME_BOT_TOKEN")

	if token == "" {
		return nil, fmt.Errorf("GROUPME_BOT_TOKEN environment variable is not set")
	}

	return &GroupMeConfig{
		BotID:   BotID,
		Token:   token,
		Timeout: 20 * time.Second,
		Host:    "api.groupme.com",
		Path:    "/v3/bots/post",
	}, nil
}

func loadAuthConfig() (*AuthConfig, error) {
	APIKey := os.Getenv("API_KEY")

	if APIKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is not set")
	}

	return &AuthConfig{
		APIKey: APIKey,
	}, nil
}
