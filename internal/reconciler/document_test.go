package reconciler

import (
	"strings"
	"testing"
	"time"

	"crowfather/internal/espn"
	"crowfather/internal/sleeper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildNFLTeamDoc_ContainsTeamInfo(t *testing.T) {
	team := espn.TeamWithRoster{
		Team: espn.Team{
			TeamID:          "1",
			Name:            "Kansas City Chiefs",
			RecordSummary:   "15-2",
			SeasonSummary:   "2024",
			StandingSummary: "1st AFC West",
		},
		Roster: []espn.Athlete{
			{DisplayName: "Patrick Mahomes", Position: "QB"},
			{DisplayName: "Travis Kelce", Position: "TE"},
		},
	}

	doc := string(buildNFLTeamDoc(team))
	assert.Contains(t, doc, "# Kansas City Chiefs")
	assert.Contains(t, doc, "15-2")
	assert.Contains(t, doc, "2024")
	assert.Contains(t, doc, "1st AFC West")
	assert.Contains(t, doc, "Patrick Mahomes")
	assert.Contains(t, doc, "QB")
	assert.Contains(t, doc, "Travis Kelce")
	assert.Contains(t, doc, "TE")
}

func TestBuildNFLTeamDoc_EmptyRoster(t *testing.T) {
	team := espn.TeamWithRoster{
		Team: espn.Team{Name: "Test Team"},
	}
	doc := string(buildNFLTeamDoc(team))
	assert.Contains(t, doc, "# Test Team")
	assert.Contains(t, doc, "## Roster")
}

func TestBuildFantasyLeagueDoc_ContainsRosters(t *testing.T) {
	ld := leagueData{
		leagueID:   "league1",
		leagueName: "Main Dynasty",
		rosters: []resolvedRoster{
			{
				ownerName: "JohnFantasy",
				players: []resolvedPlayer{
					{name: "Patrick Mahomes", position: "QB", nflTeam: "Kansas City Chiefs", nflRecord: "15-2"},
				},
			},
		},
	}

	doc := string(buildFantasyLeagueDoc(ld))
	assert.Contains(t, doc, "# Fantasy League: Main Dynasty")
	assert.Contains(t, doc, "## Team: JohnFantasy")
	assert.Contains(t, doc, "Patrick Mahomes")
	assert.Contains(t, doc, "Kansas City Chiefs")
	assert.Contains(t, doc, "15-2")
}

func TestBuildFantasyLeagueDoc_IncludesTrades(t *testing.T) {
	ld := leagueData{
		leagueID:   "l1",
		leagueName: "Test League",
		trades: []resolvedTrade{
			{
				timestamp: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
				sides: []tradeSide{
					{ownerName: "Alice", receives: []string{"Mahomes", "2026 1st Round Pick"}},
					{ownerName: "Bob", receives: []string{"Kelce"}},
				},
			},
		},
	}

	doc := string(buildFantasyLeagueDoc(ld))
	assert.Contains(t, doc, "## Recent Trades")
	assert.Contains(t, doc, "Alice")
	assert.Contains(t, doc, "Mahomes")
	assert.Contains(t, doc, "2026 1st Round Pick")
	assert.Contains(t, doc, "Bob")
	assert.Contains(t, doc, "Kelce")
}

func TestBuildTradeSummary_NoTrades(t *testing.T) {
	leagues := []leagueData{
		{leagueName: "League A", trades: nil},
	}
	summary := buildTradeSummary(leagues)
	assert.Contains(t, summary, "Rosters refreshed!")
	assert.Contains(t, summary, "No recent trades found.")
}

func TestBuildTradeSummary_WithTrades(t *testing.T) {
	leagues := []leagueData{
		{
			leagueName: "Dynasty League",
			trades: []resolvedTrade{
				{
					sides: []tradeSide{
						{ownerName: "Alice", receives: []string{"Mahomes"}},
						{ownerName: "Bob", receives: []string{"Kelce"}},
					},
				},
			},
		},
	}
	summary := buildTradeSummary(leagues)
	assert.Contains(t, summary, "Rosters refreshed!")
	assert.Contains(t, summary, "Dynasty League")
	assert.Contains(t, summary, "Alice")
	assert.Contains(t, summary, "Bob")
	assert.NotContains(t, summary, "No recent trades found.")
}

func TestResolveLeague_MapsOwnerNames(t *testing.T) {
	rosters := []sleeper.Roster{
		{RosterID: 1, OwnerID: "u1", Players: []string{"p1"}},
	}
	users := []sleeper.User{
		{UserID: "u1", DisplayName: "JohnFantasy"},
	}
	players := map[string]sleeper.SleeperPlayer{
		"p1": {PlayerID: "p1", FullName: "Patrick Mahomes", Position: "QB", Team: "KC"},
	}

	ld := resolveLeague("l1", "Test League", rosters, users, players, nil, nil)
	require.Len(t, ld.rosters, 1)
	assert.Equal(t, "JohnFantasy", ld.rosters[0].ownerName)
	require.Len(t, ld.rosters[0].players, 1)
	assert.Equal(t, "Patrick Mahomes", ld.rosters[0].players[0].name)
}

func TestResolveLeague_FallbackOwnerName(t *testing.T) {
	// Roster with no matching user → fallback to "Team N"
	rosters := []sleeper.Roster{
		{RosterID: 7, OwnerID: "unknown", Players: nil},
	}
	ld := resolveLeague("l1", "L", rosters, nil, nil, nil, nil)
	require.Len(t, ld.rosters, 1)
	assert.Equal(t, "Team 7", ld.rosters[0].ownerName)
}

func TestResolveLeague_CrossReferencesESPN(t *testing.T) {
	rosters := []sleeper.Roster{
		{RosterID: 1, OwnerID: "u1", Players: []string{"p1"}},
	}
	users := []sleeper.User{{UserID: "u1", DisplayName: "Alice"}}
	players := map[string]sleeper.SleeperPlayer{
		"p1": {PlayerID: "p1", FullName: "Patrick Mahomes", Position: "QB", Team: "KC"},
	}
	espnByName := map[string]espn.TeamWithRoster{
		"KC": {
			Team:   espn.Team{Name: "Kansas City Chiefs", RecordSummary: "15-2"},
			Roster: []espn.Athlete{{DisplayName: "Patrick Mahomes", Position: "QB"}},
		},
	}

	ld := resolveLeague("l1", "L", rosters, users, players, nil, espnByName)
	require.Len(t, ld.rosters[0].players, 1)
	p := ld.rosters[0].players[0]
	assert.Equal(t, "Kansas City Chiefs", p.nflTeam)
	assert.Equal(t, "15-2", p.nflRecord)
}

func TestResolveLeague_ResolvesCompletedTrades(t *testing.T) {
	rosters := []sleeper.Roster{
		{RosterID: 1, OwnerID: "u1"},
		{RosterID: 2, OwnerID: "u2"},
	}
	users := []sleeper.User{
		{UserID: "u1", DisplayName: "Alice"},
		{UserID: "u2", DisplayName: "Bob"},
	}
	transactions := []sleeper.Transaction{
		{
			TransactionID: "t1",
			Type:          "trade",
			Status:        "complete",
			Adds:          map[string]int{"p1": 2}, // Bob receives p1
			DraftPicks:    []sleeper.TradedPick{{Season: "2026", Round: 1, OwnerID: 1}}, // Alice receives pick
			Created:       time.Now().Unix() * 1000,
		},
	}
	players := map[string]sleeper.SleeperPlayer{
		"p1": {FullName: "Patrick Mahomes"},
	}

	ld := resolveLeague("l1", "L", rosters, users, players, transactions, nil)
	require.Len(t, ld.trades, 1)

	// Collect all items received across sides
	var allReceived []string
	for _, side := range ld.trades[0].sides {
		allReceived = append(allReceived, side.receives...)
	}
	assert.True(t, func() bool {
		for _, s := range allReceived {
			if strings.Contains(s, "Mahomes") {
				return true
			}
		}
		return false
	}(), "expected Mahomes in received items")
}

func TestOrdinal(t *testing.T) {
	cases := map[int]string{
		1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 10: "10th",
	}
	for n, want := range cases {
		assert.Equal(t, want, ordinal(n))
	}
}
