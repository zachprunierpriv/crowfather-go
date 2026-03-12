package espn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rosterJSON(teamID, teamName string, athletes []Athlete) string {
	positions := []Position{{Position: "QB", Items: athletes}}
	resp := RosterResponse{
		Team: Team{
			TeamID:          teamID,
			Name:            teamName,
			RecordSummary:   "10-7",
			SeasonSummary:   "2024",
			StandingSummary: "1st AFC West",
		},
		Athletes: positions,
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func TestFetchAllTeamRosters_ReturnsValidTeams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var teamID int
		fmt.Sscanf(r.URL.Path, "/%d", &teamID)
		if teamID == 1 {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, rosterJSON("1", "Kansas City Chiefs", []Athlete{
				{AthleteID: "1", DisplayName: "Patrick Mahomes", Position: "QB"},
			}))
			return
		}
		// All other team IDs return 404 (simulates non-existent teams).
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := &ESPNService{
		client:       server.Client(),
		rosterURLFmt: server.URL + "/%d",
	}

	teams, err := svc.FetchAllTeamRosters(context.Background())
	require.NoError(t, err)
	require.Len(t, teams, 1)
	assert.Equal(t, "Kansas City Chiefs", teams[0].Team.Name)
	assert.Equal(t, "10-7", teams[0].Team.RecordSummary)
	require.Len(t, teams[0].Roster, 1)
	assert.Equal(t, "Patrick Mahomes", teams[0].Roster[0].DisplayName)
	assert.Equal(t, "QB", teams[0].Roster[0].Position)
}

func TestFetchAllTeamRosters_AggregatesMultipleTeams(t *testing.T) {
	teams := map[int]string{1: "Chiefs", 2: "Eagles"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id int
		fmt.Sscanf(r.URL.Path, "/%d", &id)
		if name, ok := teams[id]; ok {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, rosterJSON(fmt.Sprint(id), name, nil))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := &ESPNService{client: server.Client(), rosterURLFmt: server.URL + "/%d"}
	got, err := svc.FetchAllTeamRosters(context.Background())
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestFetchAllTeamRosters_AllFailReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := &ESPNService{client: server.Client(), rosterURLFmt: server.URL + "/%d"}
	_, err := svc.FetchAllTeamRosters(context.Background())
	assert.Error(t, err)
}

func TestFetchAllTeamRosters_PositionTaggedOnAthletes(t *testing.T) {
	// Verifies that the position from the Position group is set on each Athlete.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id int
		fmt.Sscanf(r.URL.Path, "/%d", &id)
		if id != 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp := RosterResponse{
			Team: Team{TeamID: "1", Name: "Chiefs"},
			Athletes: []Position{
				{Position: "QB", Items: []Athlete{{AthleteID: "10", DisplayName: "Mahomes"}}},
				{Position: "WR", Items: []Athlete{{AthleteID: "11", DisplayName: "Rice"}}},
			},
		}
		b, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}))
	defer server.Close()

	svc := &ESPNService{client: server.Client(), rosterURLFmt: server.URL + "/%d"}
	teams, err := svc.FetchAllTeamRosters(context.Background())
	require.NoError(t, err)
	require.Len(t, teams[0].Roster, 2)

	byName := make(map[string]Athlete)
	for _, a := range teams[0].Roster {
		byName[a.DisplayName] = a
	}
	assert.Equal(t, "QB", byName["Mahomes"].Position)
	assert.Equal(t, "WR", byName["Rice"].Position)
}
