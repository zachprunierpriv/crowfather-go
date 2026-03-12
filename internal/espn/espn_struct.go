package espn

type RosterResponse struct {
	Timestamp string     `json:"timestamp"`
	Status    string     `json:"status"`
	Athletes  []Position `json:"athletes"`
	Team      Team       `json:"team"`
}

type Position struct {
	Position string    `json:"position"`
	Items    []Athlete `json:"items"`
}

type Athlete struct {
	AthleteID   string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Position    string `json:"position"`
}

type Team struct {
	TeamID          string `json:"id"`
	Name            string `json:"name"`
	RecordSummary   string `json:"recordSummary"`
	SeasonSummary   string `json:"seasonSummary"`
	StandingSummary string `json:"standingSummary"`
}

type TeamWithRoster struct {
	Team   Team      `json:"team"`
	Roster []Athlete `json:"roster"`
}
