package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type GroupMeConfig struct {
	BotID   string        `json:"bot_id"`
	Token   string        `json:"token"`
	Timeout time.Duration `json:"timeout"`
	Host    string        `json:"host"`
	Path    string        `json:"path"`
}

type Assistants struct {
	GroupMeAssistantID  string `json:"groupme_assistant_id"`
	MeltdownAssistantID string `json:"meltdown_assistant_id"`
	TestAssistantID     string `json:"test_assistant_id"`
}

type OpenAIConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

type Config struct {
	OpenAI     *OpenAIConfig     `json:"openai"`
	GroupMe    *GroupMeConfig    `json:"groupme"`
	Auth       *AuthConfig       `json:"auth"`
	Assistants *Assistants       `json:"assistants"`
	Reconciler *ReconcilerConfig `json:"reconciler"` // nil if not configured
}

type AuthConfig struct {
	APIKey string `json:"api_key"`
}

type ReconcilerConfig struct {
	LeagueIDs         []string      // SLEEPER_LEAGUE_IDS (comma-separated)
	OnStartup         bool          // RECONCILE_ON_STARTUP (default true)
	Interval          time.Duration // RECONCILE_INTERVAL_HOURS (default 168h)
	CooldownMinutes   time.Duration // RECONCILE_COOLDOWN_MINUTES (default 30m)
	ApprovedUsers     []string      // RECONCILE_APPROVED_USERS (comma-separated GroupMe user_ids)
	TransactionRounds int           // RECONCILE_TRANSACTION_ROUNDS (default 2)
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
	assistants, err := loadAssistants()

	if err != nil {
		return nil, err
	}

	reconcilerConfig := loadReconcilerConfig()

	return &Config{
		OpenAI:     openAIConfig,
		GroupMe:    groupMeConfig,
		Auth:       authConfig,
		Assistants: assistants,
		Reconciler: reconcilerConfig,
	}, nil
}

func loadOpenAIConfig() (*OpenAIConfig, error) {
	APIKey := os.Getenv("OPENAI_API_KEY")

	if APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return &OpenAIConfig{
		APIKey:  APIKey,
		BaseURL: "https://api.openai.com/v1",
		Timeout: 60 * time.Second,
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

func loadAssistants() (*Assistants, error) {
	GroupMeAssistantID := os.Getenv("GROUPME_ASSISTANT_ID")

	if GroupMeAssistantID == "" {
		return nil, fmt.Errorf("GROUPME_ASSISTANT_ID environment variable is not set")
	}

	MeltdownAssistantID := os.Getenv("MELTDOWN_ASSISTANT_ID")

	if MeltdownAssistantID == "" {
		return nil, fmt.Errorf("MELTDOWN_ASSISTANT_ID environment variable is not set")
	}

	TestAssistantID := os.Getenv("TEST_ASSISTANT_ID")

	if TestAssistantID == "" {
		return nil, fmt.Errorf("TEST_ASSISTANT_ID environment variable is not set")
	}

	return &Assistants{
		GroupMeAssistantID:  GroupMeAssistantID,
		MeltdownAssistantID: MeltdownAssistantID,
		TestAssistantID:     TestAssistantID,
	}, nil
}

// loadReconcilerConfig loads optional reconciler settings. Returns nil if no
// league IDs are configured, which disables the reconciler entirely.
func loadReconcilerConfig() *ReconcilerConfig {
	leagueIDsRaw := os.Getenv("SLEEPER_LEAGUE_IDS")
	if leagueIDsRaw == "" {
		return nil
	}

	leagueIDs := splitTrimmed(leagueIDsRaw)
	if len(leagueIDs) == 0 {
		return nil
	}

	onStartup := true
	if v := os.Getenv("RECONCILE_ON_STARTUP"); v == "false" {
		onStartup = false
	}

	intervalHours := 168
	if v := os.Getenv("RECONCILE_INTERVAL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalHours = n
		}
	}

	cooldownMinutes := 30
	if v := os.Getenv("RECONCILE_COOLDOWN_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cooldownMinutes = n
		}
	}

	transactionRounds := 2
	if v := os.Getenv("RECONCILE_TRANSACTION_ROUNDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			transactionRounds = n
		}
	}

	var approvedUsers []string
	if v := os.Getenv("RECONCILE_APPROVED_USERS"); v != "" {
		approvedUsers = splitTrimmed(v)
	}

	return &ReconcilerConfig{
		LeagueIDs:         leagueIDs,
		OnStartup:         onStartup,
		Interval:          time.Duration(intervalHours) * time.Hour,
		CooldownMinutes:   time.Duration(cooldownMinutes) * time.Minute,
		ApprovedUsers:     approvedUsers,
		TransactionRounds: transactionRounds,
	}
}

func splitTrimmed(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
