package sleeper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.sleeper.app/v1"

type SleeperService struct {
	client  *http.Client
	baseURL string
}

func NewSleeperService() *SleeperService {
	return &SleeperService{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultBaseURL,
	}
}

// FetchAllPlayers fetches the full NFL player map from Sleeper (~5MB).
// Returns a map keyed by Sleeper player ID for fast lookup.
func (s *SleeperService) FetchAllPlayers(ctx context.Context) (map[string]SleeperPlayer, error) {
	var players map[string]SleeperPlayer
	if err := s.get(ctx, "/players/nfl", &players); err != nil {
		return nil, fmt.Errorf("failed to fetch Sleeper players: %w", err)
	}
	return players, nil
}

// FetchLeague fetches basic league metadata.
func (s *SleeperService) FetchLeague(ctx context.Context, leagueID string) (*League, error) {
	var league League
	if err := s.get(ctx, fmt.Sprintf("/league/%s", leagueID), &league); err != nil {
		return nil, fmt.Errorf("failed to fetch league %s: %w", leagueID, err)
	}
	return &league, nil
}

// FetchLeagueRosters fetches all fantasy rosters for a league.
func (s *SleeperService) FetchLeagueRosters(ctx context.Context, leagueID string) ([]Roster, error) {
	var rosters []Roster
	if err := s.get(ctx, fmt.Sprintf("/league/%s/rosters", leagueID), &rosters); err != nil {
		return nil, fmt.Errorf("failed to fetch rosters for league %s: %w", leagueID, err)
	}
	return rosters, nil
}

// FetchLeagueUsers fetches all users in a league.
func (s *SleeperService) FetchLeagueUsers(ctx context.Context, leagueID string) ([]User, error) {
	var users []User
	if err := s.get(ctx, fmt.Sprintf("/league/%s/users", leagueID), &users); err != nil {
		return nil, fmt.Errorf("failed to fetch users for league %s: %w", leagueID, err)
	}
	return users, nil
}

// FetchRecentTransactions fetches completed trade transactions for a league,
// paginating through rounds 1..maxRounds (or until an empty page is returned).
// Only completed trades are returned.
func (s *SleeperService) FetchRecentTransactions(ctx context.Context, leagueID string, maxRounds int) ([]Transaction, error) {
	var all []Transaction
	for round := 1; round <= maxRounds; round++ {
		var page []Transaction
		if err := s.get(ctx, fmt.Sprintf("/league/%s/transactions/%d", leagueID, round), &page); err != nil {
			return nil, fmt.Errorf("failed to fetch transactions round %d for league %s: %w", round, leagueID, err)
		}
		if len(page) == 0 {
			break
		}
		for _, t := range page {
			if t.Type == "trade" && t.Status == "complete" {
				all = append(all, t)
			}
		}
	}
	return all, nil
}

func (s *SleeperService) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
