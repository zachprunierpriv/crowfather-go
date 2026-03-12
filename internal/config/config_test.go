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

func TestLoadReconcilerConfig_NoLeagueIDs(t *testing.T) {
	t.Setenv("SLEEPER_LEAGUE_IDS", "")
	if cfg := loadReconcilerConfig(); cfg != nil {
		t.Fatalf("expected nil config when SLEEPER_LEAGUE_IDS is empty, got %+v", cfg)
	}
}

func TestLoadReconcilerConfig_WithLeagueIDs(t *testing.T) {
	t.Setenv("SLEEPER_LEAGUE_IDS", "league1,league2")
	t.Setenv("RECONCILE_ON_STARTUP", "false")
	t.Setenv("RECONCILE_INTERVAL_HOURS", "24")
	t.Setenv("RECONCILE_COOLDOWN_MINUTES", "15")
	t.Setenv("RECONCILE_TRANSACTION_ROUNDS", "3")
	t.Setenv("RECONCILE_APPROVED_USERS", "user1,user2")

	cfg := loadReconcilerConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.LeagueIDs) != 2 || cfg.LeagueIDs[0] != "league1" || cfg.LeagueIDs[1] != "league2" {
		t.Errorf("unexpected league IDs: %v", cfg.LeagueIDs)
	}
	if cfg.OnStartup != false {
		t.Error("expected OnStartup=false")
	}
	if cfg.Interval.Hours() != 24 {
		t.Errorf("unexpected interval: %v", cfg.Interval)
	}
	if cfg.CooldownMinutes.Minutes() != 15 {
		t.Errorf("unexpected cooldown: %v", cfg.CooldownMinutes)
	}
	if cfg.TransactionRounds != 3 {
		t.Errorf("unexpected transaction rounds: %d", cfg.TransactionRounds)
	}
	if len(cfg.ApprovedUsers) != 2 || cfg.ApprovedUsers[0] != "user1" {
		t.Errorf("unexpected approved users: %v", cfg.ApprovedUsers)
	}
}

func TestLoadReconcilerConfig_Defaults(t *testing.T) {
	t.Setenv("SLEEPER_LEAGUE_IDS", "league1")
	t.Setenv("RECONCILE_ON_STARTUP", "")
	t.Setenv("RECONCILE_INTERVAL_HOURS", "")
	t.Setenv("RECONCILE_COOLDOWN_MINUTES", "")
	t.Setenv("RECONCILE_TRANSACTION_ROUNDS", "")
	t.Setenv("RECONCILE_APPROVED_USERS", "")

	cfg := loadReconcilerConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if !cfg.OnStartup {
		t.Error("expected OnStartup=true by default")
	}
	if cfg.Interval.Hours() != 168 {
		t.Errorf("expected 168h default interval, got %v", cfg.Interval)
	}
	if cfg.CooldownMinutes.Minutes() != 30 {
		t.Errorf("expected 30m default cooldown, got %v", cfg.CooldownMinutes)
	}
	if cfg.TransactionRounds != 2 {
		t.Errorf("expected 2 default transaction rounds, got %d", cfg.TransactionRounds)
	}
	if len(cfg.ApprovedUsers) != 0 {
		t.Errorf("expected empty approved users by default, got %v", cfg.ApprovedUsers)
	}
}

func TestSplitTrimmed(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b ", []string{"a", "b"}},
		{"", nil},
		{",,,", nil},
	}
	for _, tc := range cases {
		got := splitTrimmed(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("splitTrimmed(%q) = %v, want %v", tc.input, got, tc.expected)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("splitTrimmed(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.expected[i])
			}
		}
	}
}
