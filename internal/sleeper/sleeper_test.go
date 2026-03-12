package sleeper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSleeper(server *httptest.Server) *SleeperService {
	return &SleeperService{client: server.Client(), baseURL: server.URL}
}

func TestFetchAllPlayers(t *testing.T) {
	players := map[string]SleeperPlayer{
		"1234": {PlayerID: "1234", FullName: "Patrick Mahomes", Position: "QB", Team: "KC"},
		"5678": {PlayerID: "5678", FullName: "Travis Kelce", Position: "TE", Team: "KC"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/players/nfl", r.URL.Path)
		json.NewEncoder(w).Encode(players)
	}))
	defer server.Close()

	got, err := newTestSleeper(server).FetchAllPlayers(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Patrick Mahomes", got["1234"].FullName)
}

func TestFetchLeagueRosters(t *testing.T) {
	rosters := []Roster{
		{RosterID: 1, OwnerID: "user1", Players: []string{"1234", "5678"}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/league/myLeague/rosters", r.URL.Path)
		json.NewEncoder(w).Encode(rosters)
	}))
	defer server.Close()

	got, err := newTestSleeper(server).FetchLeagueRosters(context.Background(), "myLeague")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, []string{"1234", "5678"}, got[0].Players)
}

func TestFetchLeagueUsers(t *testing.T) {
	users := []User{
		{UserID: "user1", DisplayName: "JohnFantasy"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/league/myLeague/users", r.URL.Path)
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	got, err := newTestSleeper(server).FetchLeagueUsers(context.Background(), "myLeague")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "JohnFantasy", got[0].DisplayName)
}

func TestFetchRecentTransactions_ReturnsTrades(t *testing.T) {
	txs := []Transaction{
		{TransactionID: "t1", Type: "trade", Status: "complete", RosterIDs: []int{1, 2}},
		{TransactionID: "t2", Type: "waiver", Status: "complete"},   // should be filtered out
		{TransactionID: "t3", Type: "trade", Status: "pending"},     // should be filtered out
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/1") {
			json.NewEncoder(w).Encode(txs)
			return
		}
		json.NewEncoder(w).Encode([]Transaction{}) // round 2 empty → stop
	}))
	defer server.Close()

	got, err := newTestSleeper(server).FetchRecentTransactions(context.Background(), "league1", 3)
	require.NoError(t, err)
	require.Len(t, got, 1, "only completed trades should be returned")
	assert.Equal(t, "t1", got[0].TransactionID)
}

func TestFetchRecentTransactions_StopsAtEmptyPage(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if strings.HasSuffix(r.URL.Path, "/1") {
			json.NewEncoder(w).Encode([]Transaction{
				{TransactionID: "t1", Type: "trade", Status: "complete"},
			})
			return
		}
		json.NewEncoder(w).Encode([]Transaction{}) // empty on round 2
	}))
	defer server.Close()

	_, err := newTestSleeper(server).FetchRecentTransactions(context.Background(), "league1", 5)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount, "should stop after the empty page, not fetch all 5 rounds")
}

func TestFetchRecentTransactions_IncludesDraftPicks(t *testing.T) {
	txs := []Transaction{{
		TransactionID: "t1",
		Type:          "trade",
		Status:        "complete",
		DraftPicks: []TradedPick{
			{Season: "2026", Round: 1, OwnerID: 2, PreviousOwnerID: 1},
		},
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/1") {
			json.NewEncoder(w).Encode(txs)
			return
		}
		json.NewEncoder(w).Encode([]Transaction{})
	}))
	defer server.Close()

	got, err := newTestSleeper(server).FetchRecentTransactions(context.Background(), "league1", 2)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Len(t, got[0].DraftPicks, 1)
	assert.Equal(t, "2026", got[0].DraftPicks[0].Season)
}

func TestFetchLeague_ErrorOnNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := newTestSleeper(server).FetchLeague(context.Background(), "bad")
	assert.Error(t, err)
}
