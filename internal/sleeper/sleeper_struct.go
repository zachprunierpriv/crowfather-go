package sleeper

type SleeperPlayer struct {
	PlayerID string `json:"player_id"`
	FullName string `json:"full_name"`
	Position string `json:"position"`
	Team     string `json:"team"` // NFL team abbreviation, e.g. "KC"
}

type Roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"` // Sleeper player IDs
}

type User struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type League struct {
	LeagueID string `json:"league_id"`
	Name     string `json:"name"`
}

type Transaction struct {
	TransactionID string         `json:"transaction_id"`
	Type          string         `json:"type"`   // "trade", "waiver", "free_agent"
	Status        string         `json:"status"` // "complete", "pending", "cancelled"
	Adds          map[string]int `json:"adds"`   // player_id → roster_id receiving
	Drops         map[string]int `json:"drops"`  // player_id → roster_id dropping
	DraftPicks    []TradedPick   `json:"draft_picks"`
	RosterIDs     []int          `json:"roster_ids"`
	Created       int64          `json:"created"`
}

type TradedPick struct {
	Season          string `json:"season"`
	Round           int    `json:"round"`
	OwnerID         int    `json:"owner_id"`
	PreviousOwnerID int    `json:"previous_owner_id"`
}
