package espn

// rawAthletePosition matches ESPN's nested position object on each athlete.
type rawAthletePosition struct {
	Abbreviation string `json:"abbreviation"`
}

// rawAthlete is used only for decoding the ESPN API response.
// ESPN returns each athlete's position as an object, not a string.
type rawAthlete struct {
	AthleteID   string             `json:"id"`
	FirstName   string             `json:"firstName"`
	LastName    string             `json:"lastName"`
	DisplayName string             `json:"displayName"`
	Position    rawAthletePosition `json:"position"`
}

type rawPosition struct {
	Position string       `json:"position"`
	Items    []rawAthlete `json:"items"`
}

type RosterResponse struct {
	Timestamp string        `json:"timestamp"`
	Status    string        `json:"status"`
	Athletes  []rawPosition `json:"athletes"`
	Team      Team          `json:"team"`
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
	Position    string
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
