package api

// FIFATeamOption is one team entry for the FIFA teams list.
type FIFATeamOption struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
	Elo   int    `json:"elo"`
	Host  bool   `json:"host"`
}

// FIFAListTeamsResponse lists all FIFA World Cup teams.
type FIFAListTeamsResponse struct {
	Teams []FIFATeamOption `json:"teams"`
}

// FIFASimulateRequest is the body for POST /fifa/simulate.
type FIFASimulateRequest struct {
	Team1ID    string `json:"team1_id" example:"USA"`
	Team2ID    string `json:"team2_id" example:"MEX"`
	Seed       *int64 `json:"seed,omitempty" example:"42"`
	ExtraTime  bool   `json:"extra_time" example:"false"`
	HostTeamID string `json:"host_team_id,omitempty" example:"USA"` // optional: team receiving +100 Elo crowd boost
}

// FIFASimulateResponse is returned after a FIFA match simulation.
// All matches are played at neutral venues — team1/team2 are positional only.
type FIFASimulateResponse struct {
	Seed          int64              `json:"seed"`
	Team1ID       string             `json:"team1_id"`
	Team2ID       string             `json:"team2_id"`
	Team1Name     string             `json:"team1_name"`
	Team2Name     string             `json:"team2_name"`
	Team1Goals    int                `json:"team1_goals"`
	Team2Goals    int                `json:"team2_goals"`
	Team1GoalsAET *int               `json:"team1_goals_aet,omitempty"`
	Team2GoalsAET *int               `json:"team2_goals_aet,omitempty"`
	Penalties     *FIFAPenaltyResult `json:"penalties,omitempty"`
	Winner        string             `json:"winner,omitempty"` // empty = draw (group stage)
	Period        string             `json:"period"`           // "90" | "aet" | "pens"
}

// FIFAPenaltyResult holds the full penalty shootout outcome.
type FIFAPenaltyResult struct {
	Team1Scored int               `json:"team1_scored"`
	Team2Scored int               `json:"team2_scored"`
	Kicks       []FIFAPenaltyKick `json:"kicks"`
}

// FIFAPenaltyKick is one penalty attempt.
type FIFAPenaltyKick struct {
	Team   string `json:"team"`
	KickNo int    `json:"kick_no"`
	Scored bool   `json:"scored"`
}

// FIFASimulateGroupStageRequest is the body for POST /fifa/simulate/group-stage.
type FIFASimulateGroupStageRequest struct {
	Seed *int64 `json:"seed,omitempty" example:"42"`
}

// FIFATeamStanding is one team's standing within a group.
type FIFATeamStanding struct {
	TeamID  string `json:"team_id"`
	Name    string `json:"name"`
	MP      int    `json:"mp"`
	W       int    `json:"w"`
	D       int    `json:"d"`
	L       int    `json:"l"`
	GF      int    `json:"gf"`
	GA      int    `json:"ga"`
	GD      int    `json:"gd"`
	Pts     int    `json:"pts"`
	Advance bool   `json:"advance"`
}

// FIFAGroupStanding holds the final standings for one group.
type FIFAGroupStanding struct {
	Group string             `json:"group"`
	Teams []FIFATeamStanding `json:"teams"`
}

// FIFASimulateGroupStageResponse is returned after a full group-stage simulation.
type FIFASimulateGroupStageResponse struct {
	Seed   int64               `json:"seed"`
	Groups []FIFAGroupStanding `json:"groups"`
}
