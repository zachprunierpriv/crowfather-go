package reconciler

import (
	"crowfather/internal/espn"
	"crowfather/internal/sleeper"
	"fmt"
	"strings"
	"time"
)

// buildNFLTeamDoc generates a Markdown document for a single NFL team.
func buildNFLTeamDoc(team espn.TeamWithRoster) []byte {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s\n", team.Team.Name)
	if team.Team.RecordSummary != "" {
		fmt.Fprintf(&sb, "Record: %s", team.Team.RecordSummary)
	}
	if team.Team.SeasonSummary != "" {
		fmt.Fprintf(&sb, " | Season: %s", team.Team.SeasonSummary)
	}
	if team.Team.StandingSummary != "" {
		fmt.Fprintf(&sb, " | Standing: %s", team.Team.StandingSummary)
	}
	sb.WriteString("\n\n## Roster\n\n")
	sb.WriteString("| Name | Position |\n")
	sb.WriteString("|------|----------|\n")
	for _, a := range team.Roster {
		fmt.Fprintf(&sb, "| %s | %s |\n", a.DisplayName, a.Position)
	}

	return []byte(sb.String())
}

// leagueData holds resolved league information for document generation.
type leagueData struct {
	leagueID   string
	leagueName string
	rosters    []resolvedRoster
	trades     []resolvedTrade
}

type resolvedRoster struct {
	ownerName string
	players   []resolvedPlayer
}

type resolvedPlayer struct {
	name       string
	position   string
	nflTeam    string
	nflRecord  string
}

type resolvedTrade struct {
	timestamp time.Time
	sides     []tradeSide
}

type tradeSide struct {
	ownerName string
	receives  []string // player names + draft picks
}

// buildFantasyLeagueDoc generates a Markdown document for a single Sleeper league.
func buildFantasyLeagueDoc(ld leagueData) []byte {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Fantasy League: %s\n\n", ld.leagueName)

	if len(ld.trades) > 0 {
		sb.WriteString("## Recent Trades\n\n")
		for _, t := range ld.trades {
			sb.WriteString(fmt.Sprintf("**%s**\n", t.timestamp.Format("Jan 2, 2006")))
			for _, side := range t.sides {
				fmt.Fprintf(&sb, "- %s receives: %s\n", side.ownerName, strings.Join(side.receives, ", "))
			}
			sb.WriteString("\n")
		}
	}

	for _, r := range ld.rosters {
		fmt.Fprintf(&sb, "## Team: %s\n\n", r.ownerName)
		sb.WriteString("| Player | Position | NFL Team | NFL Record |\n")
		sb.WriteString("|--------|----------|----------|------------|\n")
		for _, p := range r.players {
			fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", p.name, p.position, p.nflTeam, p.nflRecord)
		}
		sb.WriteString("\n")
	}

	return []byte(sb.String())
}

// buildTradeSummary generates the GroupMe notification string for recent trades.
func buildTradeSummary(leagues []leagueData) string {
	var sb strings.Builder
	sb.WriteString("Rosters refreshed!\n\n")

	anyTrades := false
	for _, ld := range leagues {
		if len(ld.trades) == 0 {
			continue
		}
		anyTrades = true
		fmt.Fprintf(&sb, "Recent Trades - %s:\n", ld.leagueName)
		for _, t := range ld.trades {
			for i, side := range t.sides {
				if i == 0 {
					fmt.Fprintf(&sb, "  - %s", side.ownerName)
				} else {
					fmt.Fprintf(&sb, " <-> %s", side.ownerName)
				}
			}
			sb.WriteString("\n")
			for _, side := range t.sides {
				fmt.Fprintf(&sb, "    %s gets: %s\n", side.ownerName, strings.Join(side.receives, ", "))
			}
		}
		sb.WriteString("\n")
	}

	if !anyTrades {
		sb.WriteString("No recent trades found.")
	}

	return sb.String()
}

// resolveLeague converts raw Sleeper data into a leagueData struct, cross-referencing
// ESPN player data where possible.
func resolveLeague(
	leagueID string,
	leagueName string,
	rosters []sleeper.Roster,
	users []sleeper.User,
	players map[string]sleeper.SleeperPlayer,
	transactions []sleeper.Transaction,
	espnByName map[string]espn.TeamWithRoster,
) leagueData {
	// Build roster_id → owner name lookup
	ownerByRosterID := make(map[int]string)
	userByID := make(map[string]sleeper.User)
	for _, u := range users {
		userByID[u.UserID] = u
	}
	for _, r := range rosters {
		if u, ok := userByID[r.OwnerID]; ok {
			ownerByRosterID[r.RosterID] = u.DisplayName
		} else {
			ownerByRosterID[r.RosterID] = fmt.Sprintf("Team %d", r.RosterID)
		}
	}

	// Resolve rosters
	var resolved []resolvedRoster
	for _, r := range rosters {
		owner := ownerByRosterID[r.RosterID]
		var rPlayers []resolvedPlayer
		for _, pid := range r.Players {
			sp, ok := players[pid]
			if !ok {
				continue
			}
			rp := resolvedPlayer{
				name:     sp.FullName,
				position: sp.Position,
				nflTeam:  sp.Team,
			}
			// Cross-reference with ESPN for team record
			if teamData, found := espnByName[normalizeTeam(sp.Team)]; found {
				rp.nflTeam = teamData.Team.Name
				rp.nflRecord = teamData.Team.RecordSummary
			}
			rPlayers = append(rPlayers, rp)
		}
		resolved = append(resolved, resolvedRoster{ownerName: owner, players: rPlayers})
	}

	// Resolve trades
	var trades []resolvedTrade
	for _, t := range transactions {
		if t.Type != "trade" || t.Status != "complete" {
			continue
		}

		// Group adds by the receiving roster
		receives := make(map[int][]string)
		for pid, rosterID := range t.Adds {
			name := pid
			if sp, ok := players[pid]; ok {
				name = sp.FullName
			}
			receives[rosterID] = append(receives[rosterID], name)
		}
		// Add draft picks to the receiving owner
		for _, pick := range t.DraftPicks {
			pickStr := fmt.Sprintf("%s %s Round Pick", pick.Season, ordinal(pick.Round))
			receives[pick.OwnerID] = append(receives[pick.OwnerID], pickStr)
		}

		var sides []tradeSide
		for rosterID, items := range receives {
			sides = append(sides, tradeSide{
				ownerName: ownerByRosterID[rosterID],
				receives:  items,
			})
		}

		trades = append(trades, resolvedTrade{
			timestamp: time.Unix(t.Created/1000, 0),
			sides:     sides,
		})
	}

	return leagueData{
		leagueID:   leagueID,
		leagueName: leagueName,
		rosters:    resolved,
		trades:     trades,
	}
}

// normalizeTeam builds a lookup key from a Sleeper NFL team abbreviation.
// ESPN team data is stored by abbreviation in the espnByName map.
func normalizeTeam(abbrev string) string {
	return strings.ToUpper(strings.TrimSpace(abbrev))
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}
