package reconciler

import (
	"context"
	"crowfather/internal/espn"
	"crowfather/internal/open_ai"
	"crowfather/internal/sleeper"
	"fmt"
	"sync"
	"time"
)

const vectorStoreName = "crowfather-sports-data"
const vectorStoreIDKey = "vector_store_id"

// MetadataRepository is the persistence contract for key/value metadata.
// The concrete implementation lives in the database package.
type MetadataRepository interface {
	GetMetadata(ctx context.Context, key string) (string, error)
	SetMetadata(ctx context.Context, key, value string) error
}

// Reconciler orchestrates fetching ESPN + Sleeper data, generating documents,
// and uploading them to an OpenAI vector store.
type Reconciler struct {
	espn          *espn.ESPNService
	sleeper       *sleeper.SleeperService
	oai           *open_ai.OpenAIService
	db            MetadataRepository
	leagueIDs     []string
	assistantID   string
	transRounds   int
	approvedUsers map[string]bool

	mu        sync.Mutex
	running   bool
	lastRunAt time.Time
	cooldown  time.Duration
}

func NewReconciler(
	espnSvc *espn.ESPNService,
	sleeperSvc *sleeper.SleeperService,
	oai *open_ai.OpenAIService,
	db MetadataRepository,
	leagueIDs []string,
	assistantID string,
	transRounds int,
	cooldown time.Duration,
	approvedUsers []string,
) *Reconciler {
	approved := make(map[string]bool, len(approvedUsers))
	for _, u := range approvedUsers {
		approved[u] = true
	}

	return &Reconciler{
		espn:          espnSvc,
		sleeper:       sleeperSvc,
		oai:           oai,
		db:            db,
		leagueIDs:     leagueIDs,
		assistantID:   assistantID,
		transRounds:   transRounds,
		cooldown:      cooldown,
		approvedUsers: approved,
	}
}

// Trigger attempts to start a reconciliation run. senderUserID is the GroupMe
// user_id of the person triggering via chat, or "" for HTTP/cron/startup triggers.
// notify is an optional callback called with the trade summary when the run completes.
// Returns (true, "") if the run was started, or (false, reason) if blocked.
func (r *Reconciler) Trigger(senderUserID string, notify func(string)) (triggered bool, reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Access control: only applies to GroupMe triggers with a non-empty user ID.
	if senderUserID != "" && len(r.approvedUsers) > 0 {
		if !r.approvedUsers[senderUserID] {
			return false, "You're not authorized to trigger a roster refresh."
		}
	}

	// Cooldown check.
	if !r.lastRunAt.IsZero() && time.Since(r.lastRunAt) < r.cooldown {
		remaining := (r.cooldown - time.Since(r.lastRunAt)).Round(time.Minute)
		return false, fmt.Sprintf("Roster data was just refreshed. Try again in %v.", remaining)
	}

	// In-flight guard.
	if r.running {
		return false, "A roster refresh is already in progress."
	}

	r.running = true
	go func() {
		var summary string
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Printf("reconciler: recovered from panic: %v\n", rec)
				summary = fmt.Sprintf("Roster refresh failed unexpectedly: %v", rec)
				if notify != nil {
					notify(summary)
				}
			}
			r.mu.Lock()
			r.running = false
			r.lastRunAt = time.Now()
			r.mu.Unlock()
		}()
		var err error
		summary, err = r.run(context.Background())
		if err != nil {
			fmt.Printf("reconciler: run failed: %v\n", err)
			summary = fmt.Sprintf("Roster refresh failed: %v", err)
		}
		if notify != nil {
			notify(summary)
		}
	}()

	return true, ""
}

// run performs the full reconciliation cycle and returns a trade summary string.
func (r *Reconciler) run(ctx context.Context) (string, error) {
	fmt.Println("reconciler: starting data fetch")

	// 1. Fetch ESPN rosters.
	nflTeams, err := r.espn.FetchAllTeamRosters(ctx)
	if err != nil {
		return "", fmt.Errorf("espn fetch failed: %w", err)
	}
	fmt.Printf("reconciler: fetched %d NFL teams from ESPN\n", len(nflTeams))

	// Build ESPN lookup map: normalized team abbreviation → TeamWithRoster.
	// ESPN doesn't use abbreviations in its API; we build a name-based lookup.
	espnByName := make(map[string]espn.TeamWithRoster, len(nflTeams))
	for _, t := range nflTeams {
		espnByName[normalizeTeam(t.Team.Name)] = t
	}

	// 2. Fetch Sleeper all-players (large, in-memory only during this run).
	sleeperPlayers, err := r.sleeper.FetchAllPlayers(ctx)
	if err != nil {
		return "", fmt.Errorf("sleeper players fetch failed: %w", err)
	}
	fmt.Printf("reconciler: fetched %d Sleeper players\n", len(sleeperPlayers))

	// 3. Per-league: fetch rosters, users, transactions.
	var leagues []leagueData
	for _, lid := range r.leagueIDs {
		league, err := r.fetchLeagueData(ctx, lid, sleeperPlayers, espnByName)
		if err != nil {
			fmt.Printf("reconciler: skipping league %s: %v\n", lid, err)
			continue
		}
		leagues = append(leagues, league)
	}

	// 4. Generate documents.
	docs := make(map[string][]byte)
	for _, team := range nflTeams {
		key := fmt.Sprintf("nfl_team_%s.md", team.Team.TeamID)
		docs[key] = buildNFLTeamDoc(team)
	}
	for _, ld := range leagues {
		key := fmt.Sprintf("fantasy_league_%s.md", ld.leagueID)
		docs[key] = buildFantasyLeagueDoc(ld)
	}
	fmt.Printf("reconciler: generated %d documents\n", len(docs))

	// 5. Create new vector store.
	vsID, err := r.oai.CreateVectorStore(ctx, vectorStoreName)
	if err != nil {
		return "", fmt.Errorf("vector store creation failed: %w", err)
	}
	fmt.Printf("reconciler: created vector store %s\n", vsID)

	// 6. Upload documents.
	if err := r.oai.UploadFilesToVectorStore(ctx, vsID, docs); err != nil {
		return "", fmt.Errorf("vector store upload failed: %w", err)
	}
	fmt.Println("reconciler: files uploaded to vector store")

	// 7. Attach vector store to GroupMe assistant.
	if err := r.oai.AttachVectorStoreToAssistant(ctx, r.assistantID, vsID); err != nil {
		return "", fmt.Errorf("vector store attachment failed: %w", err)
	}
	fmt.Printf("reconciler: attached vector store to assistant %s\n", r.assistantID)

	// 8. Delete old vector store (if any).
	if r.db != nil {
		oldVsID, _ := r.db.GetMetadata(ctx, vectorStoreIDKey)
		if oldVsID != "" && oldVsID != vsID {
			if err := r.oai.DeleteVectorStore(ctx, oldVsID); err != nil {
				fmt.Printf("reconciler: failed to delete old vector store %s: %v\n", oldVsID, err)
			}
		}

		// 9. Persist new vector store ID.
		if err := r.db.SetMetadata(ctx, vectorStoreIDKey, vsID); err != nil {
			fmt.Printf("reconciler: failed to persist vector store ID: %v\n", err)
		}
	}

	return buildTradeSummary(leagues), nil
}

func (r *Reconciler) fetchLeagueData(
	ctx context.Context,
	leagueID string,
	sleeperPlayers map[string]sleeper.SleeperPlayer,
	espnByName map[string]espn.TeamWithRoster,
) (leagueData, error) {
	league, err := r.sleeper.FetchLeague(ctx, leagueID)
	if err != nil {
		return leagueData{}, err
	}

	rosters, err := r.sleeper.FetchLeagueRosters(ctx, leagueID)
	if err != nil {
		return leagueData{}, err
	}

	users, err := r.sleeper.FetchLeagueUsers(ctx, leagueID)
	if err != nil {
		return leagueData{}, err
	}

	transactions, err := r.sleeper.FetchRecentTransactions(ctx, leagueID, r.transRounds)
	if err != nil {
		fmt.Printf("reconciler: failed to fetch transactions for league %s: %v\n", leagueID, err)
		transactions = nil // non-fatal
	}

	return resolveLeague(leagueID, league.Name, rosters, users, sleeperPlayers, transactions, espnByName), nil
}
