package api

// TeamRecord holds the actual end-of-season W/L for one team.
type TeamRecord struct {
	ID string `json:"id"`
	W  int    `json:"w"`
	L  int    `json:"l"`
}

// SeasonInfo describes one available season and its historical standings (if known).
type SeasonInfo struct {
	Season    string       `json:"season"`
	Standings []TeamRecord `json:"standings,omitempty"`
}

// ListSeasonsResponse lists available season rating files.
type ListSeasonsResponse struct {
	Seasons       []SeasonInfo `json:"seasons"`
	DefaultSeason string       `json:"default_season"`
}

// SimulateRequest is the body for POST /simulate.
type SimulateRequest struct {
	HomeTeamID string  `json:"home_team_id" example:"LAL"`
	AwayTeamID string  `json:"away_team_id" example:"BOS"`
	Seed       *int64  `json:"seed,omitempty" example:"42"`
	Season     *string `json:"season,omitempty" example:"2024-25"`
}

// SimulateResponse is returned after a full game simulation.
type SimulateResponse struct {
	Seed      int64       `json:"seed" example:"42"`
	Truncated bool        `json:"truncated" example:"false"`
	State     GameState   `json:"state"`
	Events    []GameEvent `json:"events"`
}

// GameState is the final state of a simulated game.
type GameState struct {
	HomeID     string `json:"home_id" example:"LAL"`
	AwayID     string `json:"away_id" example:"BOS"`
	HomeName   string `json:"home_name" example:"Lakers"`
	AwayName   string `json:"away_name" example:"Celtics"`
	HomeScore  int    `json:"home_score" example:"105"`
	AwayScore  int    `json:"away_score" example:"98"`
	Offense    string `json:"offense,omitempty" example:"home"` // omitted when status is final
	Possession int    `json:"possession" example:"201"`
	Quarter    int    `json:"quarter" example:"4"`
	Period     string `json:"period" example:"Q4"`
	ClockSec   int    `json:"clock_sec" example:"0"`
	Status     string `json:"status" example:"final"`
}

// GameEvent describes a single possession.
type GameEvent struct {
	Possession int    `json:"possession" example:"1"`
	Period     string `json:"period" example:"Q4"`
	ClockSec   int    `json:"clock_sec" example:"45"`
	Team       string `json:"team" example:"home"`
	Type       string `json:"type" example:"make_2pt"`
	Points     int    `json:"points" example:"2"`
	Text       string `json:"text" example:"Lakers make_2pt (2 pts)"`
	HomeScore  int    `json:"home_score" example:"105"`
	AwayScore  int    `json:"away_score" example:"98"`
}

// TeamOption is a team entry for dropdowns.
type TeamOption struct {
	ID         string `json:"id" example:"LAL"`
	Name       string `json:"name" example:"Lakers"`
	Conference string `json:"conference" example:"west"`
	Division   string `json:"division" example:"pacific"`
}

// ListTeamsResponse lists all teams from config.
type ListTeamsResponse struct {
	Teams []TeamOption `json:"teams"`
}

// SimulateSeasonRequest is the body for POST /simulate/season.
type SimulateSeasonRequest struct {
	Seed   *int64  `json:"seed,omitempty" example:"42"`
	Season *string `json:"season,omitempty" example:"2024-25"`
}

// CupGame is the outcome of one NBA Cup knockout game.
type CupGame struct {
	Home      string `json:"home"`
	Away      string `json:"away"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
	Winner    string `json:"winner"`
	Counted   bool   `json:"counted"`
}

// ConferenceCup holds the bracket and results for one conference.
type ConferenceCup struct {
	Seeds [4]string  `json:"seeds"` // index 0 = seed 1, index 3 = wildcard
	QF    [2]CupGame `json:"qf"`    // QF[0]=1v4, QF[1]=2v3
	SF    CupGame    `json:"sf"`
}

// CupBracket is the complete NBA Cup knockout result.
type CupBracket struct {
	East  ConferenceCup `json:"east"`
	West  ConferenceCup `json:"west"`
	Final CupGame       `json:"final"`
}

// SimulateSeasonResponse is returned after a full season simulation.
type SimulateSeasonResponse struct {
	Seed   int64            `json:"seed"`
	Season string           `json:"season"`
	East   []TeamSeasonStat `json:"east"`
	West   []TeamSeasonStat `json:"west"`
	Cup    CupBracket       `json:"cup"`
}

// TeamSeasonStat holds end-of-season stats for one team.
type TeamSeasonStat struct {
	TeamID     string  `json:"team_id"      example:"LAL"`
	TeamName   string  `json:"team_name"    example:"Lakers"`
	Conference string  `json:"conference"   example:"west"`
	Division   string  `json:"division"     example:"pacific"`
	W          int     `json:"w"            example:"54"`
	L          int     `json:"l"            example:"28"`
	ConfRecord string  `json:"conf_record"  example:"32-20"`
	DivRecord  string  `json:"div_record"   example:"10-4"`
	Streak     string  `json:"streak"       example:"W3"`
	Last10     string  `json:"last_10"      example:"7-3"`
	HomeRecord string  `json:"home_record"  example:"30-11"`
	AwayRecord string  `json:"away_record"  example:"24-17"`
	PPG        float64 `json:"ppg"          example:"114.2"`
	OPPG       float64 `json:"oppg"         example:"109.8"`
	Diff       float64 `json:"diff"         example:"4.4"`
}

// PlayInTeams holds the four play-in team IDs for one conference.
type PlayInTeams struct {
	Seed7  string `json:"seed7" example:"MIL"`
	Seed8  string `json:"seed8" example:"MIA"`
	Seed9  string `json:"seed9" example:"CHI"`
	Seed10 string `json:"seed10" example:"ATL"`
}

// SimulatePlayInRequest is the body for POST /simulate/playin.
type SimulatePlayInRequest struct {
	East   PlayInTeams `json:"east"`
	West   PlayInTeams `json:"west"`
	Seed   *int64      `json:"seed,omitempty" example:"42"`
	Season *string     `json:"season,omitempty" example:"2024-25"`
}

// PlayInGame is the result of one play-in game.
type PlayInGame struct {
	Home      string `json:"home"`
	Away      string `json:"away"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
	Winner    string `json:"winner"`
	Loser     string `json:"loser"`
}

// ConferencePlayIn holds the three play-in games and playoff seeds for one conference.
type ConferencePlayIn struct {
	Game1    PlayInGame `json:"game1"`
	Game2    PlayInGame `json:"game2"`
	Game3    PlayInGame `json:"game3"`
	Playoff7 string     `json:"playoff_7"`
	Playoff8 string     `json:"playoff_8"`
}

// SimulatePlayInResponse is returned after a play-in simulation.
type SimulatePlayInResponse struct {
	Seed int64            `json:"seed"`
	East ConferencePlayIn `json:"east"`
	West ConferencePlayIn `json:"west"`
}

// SimulateDraftLotteryRequest is the body for POST /simulate/draft-lottery.
// Teams must contain exactly 14 team IDs ordered by lottery seed:
// index 0 = seed 1 (worst record), index 13 = seed 14.
type SimulateDraftLotteryRequest struct {
	Teams [14]string `json:"teams"`
	Seed  *int64     `json:"seed,omitempty" example:"42"`
}

// DraftPick maps a draft pick number to the team that received it.
type DraftPick struct {
	Pick   int    `json:"pick" example:"1"`
	TeamID string `json:"team_id" example:"DET"`
	Seed   int    `json:"seed" example:"1"`
}

// SimulateDraftLotteryResponse is returned after a draft lottery simulation.
type SimulateDraftLotteryResponse struct {
	Seed  int64       `json:"seed" example:"42"`
	Picks []DraftPick `json:"picks"`
}

// PlayoffTeamInput is one seeded playoff team with regular-season record.
type PlayoffTeamInput struct {
	Seed    int    `json:"seed" example:"1"`
	TeamID  string `json:"team_id" example:"BOS"`
	SeasonW int    `json:"season_w" example:"60"`
	SeasonL int    `json:"season_l" example:"22"`
}

// PlayoffSeriesInput is the optional partial state of one playoff series.
// HomeTeam/AwayTeam are derived from seeding if empty.
type PlayoffSeriesInput struct {
	HomeTeam string `json:"home_team,omitempty" example:"BOS"`
	AwayTeam string `json:"away_team,omitempty" example:"MIA"`
	HomeWins int    `json:"home_wins" example:"2"`
	AwayWins int    `json:"away_wins" example:"1"`
}

// PlayoffBracketInput is the optional partial bracket state.
// Omit or zero-fill any series that hasn't started.
type PlayoffBracketInput struct {
	EastR1 [4]PlayoffSeriesInput `json:"east_r1"`
	EastR2 [2]PlayoffSeriesInput `json:"east_r2"`
	EastR3 PlayoffSeriesInput    `json:"east_r3"`
	WestR1 [4]PlayoffSeriesInput `json:"west_r1"`
	WestR2 [2]PlayoffSeriesInput `json:"west_r2"`
	WestR3 PlayoffSeriesInput    `json:"west_r3"`
	Finals PlayoffSeriesInput    `json:"finals"`
}

// SimulatePlayoffsRequest is the body for POST /nba/simulate/playoffs.
type SimulatePlayoffsRequest struct {
	East    []PlayoffTeamInput   `json:"east"`
	West    []PlayoffTeamInput   `json:"west"`
	Bracket *PlayoffBracketInput `json:"bracket,omitempty"`
	Seed    *int64               `json:"seed,omitempty" example:"42"`
	Season  *string              `json:"season,omitempty" example:"2024-25"`
}

// PlayoffGame is the result of one playoff game.
type PlayoffGame struct {
	GameNum   int    `json:"game_num" example:"1"`
	Home      string `json:"home" example:"BOS"`
	Away      string `json:"away" example:"MIA"`
	HomeScore int    `json:"home_score" example:"105"`
	AwayScore int    `json:"away_score" example:"98"`
	Winner    string `json:"winner" example:"BOS"`
}

// PlayoffSeries is the result of one playoff series.
type PlayoffSeries struct {
	HomeTeam string        `json:"home_team" example:"BOS"`
	AwayTeam string        `json:"away_team" example:"MIA"`
	HomeWins int           `json:"home_wins" example:"4"`
	AwayWins int           `json:"away_wins" example:"2"`
	Winner   string        `json:"winner" example:"BOS"`
	Games    []PlayoffGame `json:"games"`
}

// ConferencePlayoffBracket holds the bracket results for one conference.
type ConferencePlayoffBracket struct {
	R1       [4]PlayoffSeries `json:"r1"`
	R2       [2]PlayoffSeries `json:"r2"`
	R3       PlayoffSeries    `json:"r3"`
	Champion string           `json:"champion" example:"BOS"`
}

// SimulatePlayoffsResponse is returned after a playoff simulation.
type SimulatePlayoffsResponse struct {
	Seed     int64                    `json:"seed" example:"42"`
	East     ConferencePlayoffBracket `json:"east"`
	West     ConferencePlayoffBracket `json:"west"`
	Finals   PlayoffSeries            `json:"finals"`
	Champion string                   `json:"champion" example:"BOS"`
}

// ErrorResponse is returned on 4xx errors.
type ErrorResponse struct {
	Error string `json:"error" example:"unknown team id"`
}
