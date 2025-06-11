package config

import "testing"

func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("GROUPME_BOT_ID", "bot")
	t.Setenv("GROUPME_BOT_TOKEN", "token")
	t.Setenv("API_KEY", "secret")
	t.Setenv("GROUPME_ASSISTANT_ID", "gmaid")
	t.Setenv("MELTDOWN_ASSISTANT_ID", "meltdown")
	t.Setenv("TEST_ASSISTANT_ID", "test")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.OpenAI.APIKey != "openai" {
		t.Errorf("unexpected openai api key: %s", cfg.OpenAI.APIKey)
	}
	if cfg.GroupMe.BotID != "bot" {
		t.Errorf("unexpected bot id: %s", cfg.GroupMe.BotID)
	}
	if cfg.Auth.APIKey != "secret" {
		t.Errorf("unexpected api key: %s", cfg.Auth.APIKey)
	}
}

func TestLoadOpenAIConfigMissing(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	if _, err := loadOpenAIConfig(); err == nil {
		t.Fatalf("expected error when OPENAI_API_KEY missing")
	}
}
