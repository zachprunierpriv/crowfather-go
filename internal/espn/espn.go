package espn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const defaultRosterURLFmt = "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams/%d/roster"

// maxTeamID: ESPN team IDs run 1–34 (some may be unused); we skip empty responses.
const maxTeamID = 34

type ESPNService struct {
	client       *http.Client
	rosterURLFmt string // format string with one %d for team ID
}

func NewESPNService() *ESPNService {
	return &ESPNService{
		client:       &http.Client{Timeout: 15 * time.Second},
		rosterURLFmt: defaultRosterURLFmt,
	}
}

func (s *ESPNService) FetchAllTeamRosters(ctx context.Context) ([]TeamWithRoster, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	type result struct {
		roster *TeamWithRoster
		err    error
	}

	results := make(chan result, maxTeamID)
	var wg sync.WaitGroup

	for i := 1; i <= maxTeamID; i++ {
		wg.Add(1)
		go func(teamID int) {
			defer func() {
				if rec := recover(); rec != nil {
					results <- result{err: fmt.Errorf("panic fetching team %d: %v", teamID, rec)}
				}
				wg.Done()
			}()
			twr, err := s.fetchTeamRoster(ctx, teamID)
			results <- result{roster: twr, err: err}
		}(i)
	}

	wg.Wait()
	close(results)

	var teams []TeamWithRoster
	for r := range results {
		if r.err != nil || r.roster == nil {
			continue
		}
		teams = append(teams, *r.roster)
	}

	if len(teams) == 0 {
		return nil, fmt.Errorf("failed to fetch any ESPN team rosters")
	}

	return teams, nil
}

func (s *ESPNService) fetchTeamRoster(ctx context.Context, teamID int) (*TeamWithRoster, error) {
	url := fmt.Sprintf(s.rosterURLFmt, teamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil // skip invalid/non-existent team IDs silently
	}

	var rr RosterResponse
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return nil, err
	}

	if rr.Team.TeamID == "" {
		return nil, nil
	}

	twr := &TeamWithRoster{Team: rr.Team}
	for _, pos := range rr.Athletes {
		for _, athlete := range pos.Items {
			athlete.Position = pos.Position
			twr.Roster = append(twr.Roster, athlete)
		}
	}

	return twr, nil
}
