package gsis

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Signalr doesn't know what an int is. It just uses strings for everything. -_-
type StringInt int

func (si *StringInt) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case float64:
		*si = StringInt(v)
		return nil
	case int:
		*si = StringInt(v)
		return nil
	case string:
		if v == "" {
			*si = 0
		} else if n, err := strconv.Atoi(v); err != nil {
			return fmt.Errorf("error unmarshaling string int: %w", err)
		} else {
			*si = StringInt(n)
		}
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("unexpected string int type: %T", v)
	}
}

// Signalr doesn't know what a float is. It just uses strings for everything. -_-
type StringFloat float64

func (si *StringFloat) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case float64:
		*si = StringFloat(v)
		return nil
	case int:
		*si = StringFloat(v)
		return nil
	case string:
		if v == "" {
			*si = 0
		} else if n, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("error unmarshaling string float: %w", err)
		} else {
			*si = StringFloat(n)
		}
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("unexpected string float type: %T", v)
	}
}

// Signalr doesn't know what a boolean is. It just uses strings for everything. -_-
type StringBool bool

func (sb *StringBool) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case int:
		*sb = StringBool(v != 0)
		return nil
	case string:
		*sb = strings.ToLower(v) != "false" && v != ""
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("unexpected string bool type: %T", v)
	}
}

type StatFile struct {
	CumulativeStatisticsFile *StatFileCumulativeStatisticsFile
	CumeStatHeader           *StatFileCumeStatHeader
	Play                     []*StatFilePlay
	PlayStat                 []StatFilePlayStat
	PlayStatNullified        []StatFilePlayStat
	HomeTeamStats            *StatFileTeamStats
	FieldGoals               *StatFileFieldGoals
	GameAttributes           *StatFileGameAttributes
	VisitorTeamStats         *StatFileTeamStats
	ScoringSummary           []*StatFileScoringSummaryEvent
	Punts                    *StatFilePunts
	XMLName                  struct{} `xml:"CumulativeStatisticsFile" json:"-"`
}

// Update updates the StatFile based on a new one. This is useful for GSIS's incremental STATXML
// API.
func (f *StatFile) Update(update *StatFile) {
	// for the most part the contents of the file are cumulative
	f.CumeStatHeader = update.CumeStatHeader
	f.ScoringSummary = update.ScoringSummary
	f.HomeTeamStats = update.HomeTeamStats
	f.VisitorTeamStats = update.VisitorTeamStats
	f.GameAttributes = update.GameAttributes
	f.FieldGoals = update.FieldGoals
	f.Punts = update.Punts

	if len(update.Play) > 1 {
		f.Play = update.Play
		f.PlayStat = update.PlayStat
		f.PlayStatNullified = update.PlayStatNullified
	} else if len(update.Play) == 1 {
		p := update.Play[0]

		// delete the play if necessary
		if p.PlayDeleted != 0 {
			for i, existing := range f.Play {
				if existing.PlayID == p.PlayID {
					f.Play = append(f.Play[:i], f.Play[i+1:]...)
					break
				}
			}

			n := 0
			for _, existing := range f.PlayStat {
				if existing.PlayID != p.PlayID {
					f.PlayStat[n] = existing
					n += 1
				}
			}
			f.PlayStat = f.PlayStat[:n]

			n = 0
			for _, existing := range f.PlayStatNullified {
				if existing.PlayID != p.PlayID {
					f.PlayStatNullified[n] = existing
					n += 1
				}
			}
			f.PlayStatNullified = f.PlayStatNullified[:n]

			return
		}

		// if this play previously existed we need to replace it
		existed := false
		for i, existing := range f.Play {
			if existing.PlayID == p.PlayID {
				f.Play[i] = p
				existed = true
				break
			}
		}
		if existed {
			n := 0
			for _, existing := range f.PlayStat {
				if existing.PlayID != p.PlayID {
					f.PlayStat[n] = existing
					n += 1
				}
			}
			f.PlayStat = f.PlayStat[:n]
			f.PlayStat = append(f.PlayStat, update.PlayStat...)

			n = 0
			for _, existing := range f.PlayStatNullified {
				if existing.PlayID != p.PlayID {
					f.PlayStatNullified[n] = existing
					n += 1
				}
			}
			f.PlayStatNullified = f.PlayStatNullified[:n]
			f.PlayStatNullified = append(f.PlayStatNullified, update.PlayStatNullified...)
		} else {
			f.Play = append(f.Play, p)
			f.PlayStat = append(f.PlayStat, update.PlayStat...)
			f.PlayStatNullified = append(f.PlayStatNullified, update.PlayStatNullified...)
		}
		sort.Slice(f.Play, func(i, j int) bool {
			return f.Play[i].PlaySeq < f.Play[j].PlaySeq
		})
		sort.SliceStable(f.PlayStat, func(i, j int) bool {
			return f.PlayStat[i].PlayID < f.PlayStat[j].PlayID
		})
		sort.SliceStable(f.PlayStatNullified, func(i, j int) bool {
			return f.PlayStatNullified[i].PlayID < f.PlayStatNullified[j].PlayID
		})
	}
}

type StatFileCumulativeStatisticsFile struct {
	DateTimeStampUTC time.Time `xml:",attr"`
}

// A string naming one or more officials. Examples:
//
// - "Nathan Jones (42)"
// - "Walker, Jabir (26)"
// - "Frantz, Earnie & Valenti, Terri"
type StatFileOfficials string

type StatFileGameAttributes struct {
	Referee        StatFileOfficials `xml:",attr"`
	Umpire         StatFileOfficials `xml:",attr"`
	HeadLinesman   StatFileOfficials `xml:",attr"`
	DownJudge      StatFileOfficials `xml:",attr"`
	LineJudge      StatFileOfficials `xml:",attr"`
	FieldJudge     StatFileOfficials `xml:",attr"`
	SideJudge      StatFileOfficials `xml:",attr"`
	BackJudge      StatFileOfficials `xml:",attr"`
	ReplayOfficial StatFileOfficials `xml:",attr"`
}

func (a StatFileGameAttributes) Officials() []StatFileOfficial {
	ret := []StatFileOfficial{}
	for _, officials := range []StatFileOfficials{
		a.Referee,
		a.Umpire,
		a.HeadLinesman,
		a.DownJudge,
		a.LineJudge,
		a.FieldJudge,
		a.SideJudge,
		a.BackJudge,
		a.ReplayOfficial,
	} {
		ret = append(ret, officials.Officials()...)
	}
	return ret
}

type StatFileOfficial struct {
	FirstName    string
	LastName     string
	JerseyNumber string
}

var officialNameJerseyNumberRegexp = regexp.MustCompile(`^([^(]*)(?:\(([^)]*)\))?`)
var nameSuffixRegexp = regexp.MustCompile(`\s*\b([IVX]+|Sr|Jr).?\s*$`)

// Makes a best-effort attempt to parse the officials. Because there's no well-defined format for
// these, errors are ignored.
func (o StatFileOfficials) Officials() []StatFileOfficial {
	var ret []StatFileOfficial
	for _, s := range strings.Split(string(o), "&") {
		s := strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		matches := officialNameJerseyNumberRegexp.FindStringSubmatch(s)
		official := StatFileOfficial{
			JerseyNumber: matches[2],
		}
		if parts := strings.Split(matches[1], ","); len(parts) > 1 {
			// last name first
			official.LastName = strings.Fields(strings.TrimSpace(parts[0]))[0]
			official.FirstName = strings.Fields(strings.TrimSpace(parts[1]))[0]
		} else {
			// first name first
			parts := strings.Fields(nameSuffixRegexp.ReplaceAllString(strings.TrimSpace(parts[0]), ""))
			official.FirstName = parts[0]
			if len(parts) > 0 {
				official.LastName = parts[len(parts)-1]
			}
		}
		ret = append(ret, official)
	}
	return ret
}

type StatFileTeamStats struct {
	HomeTeam                               string    `xml:",attr"`
	VisitingTeam                           string    `xml:",attr"`
	DefensiveTwoPointConversions           StringInt `xml:",attr"`
	ExtraPointKickingAttempts              StringInt `xml:",attr"`
	ExtraPointKickingBlocked               StringInt `xml:",attr"`
	ExtraPointKickingSuccesses             StringInt `xml:",attr"`
	FirstDownsByPenalty                    StringInt `xml:",attr"`
	Fumbles                                StringInt `xml:",attr"`
	GoalToGoAttempts                       StringInt `xml:",attr"`
	GoalToGoSuccesses                      StringInt `xml:",attr"`
	Interceptions                          StringInt `xml:",attr"`
	InterceptionsReturnYards               StringInt `xml:",attr"`
	InterceptionsReturned                  StringInt `xml:",attr"`
	Kickoffs                               StringInt `xml:",attr"`
	KickoffsInEndZone                      StringInt `xml:",attr"`
	KickoffsReturnYards                    StringInt `xml:",attr"`
	KickoffsReturned                       StringInt `xml:",attr"`
	KickoffsTouchbacks                     StringInt `xml:",attr"`
	LostFumbles                            StringInt `xml:",attr"`
	OTScore                                StringInt `xml:",attr"`
	OnePointSafeties                       StringInt `xml:",attr"`
	PassingAttempts                        StringInt `xml:",attr"`
	PassingCompletions                     StringInt `xml:",attr"`
	PassingFirstDowns                      StringInt `xml:",attr"`
	PassingTDs                             StringInt `xml:",attr"`
	PassingYards                           StringInt `xml:",attr"`
	Penalties                              StringInt `xml:",attr"`
	PenaltyYards                           StringInt `xml:",attr"`
	Q1Score                                StringInt `xml:",attr"`
	Q2Score                                StringInt `xml:",attr"`
	Q3Score                                StringInt `xml:",attr"`
	Q4Score                                StringInt `xml:",attr"`
	RedZoneAttempts                        StringInt `xml:",attr"`
	RedZoneSuccesses                       StringInt `xml:",attr"`
	RushingFirstDowns                      StringInt `xml:",attr"`
	RushingPlays                           StringInt `xml:",attr"`
	RushingTDs                             StringInt `xml:",attr"`
	RushingYards                           StringInt `xml:",attr"`
	Safeties                               StringInt `xml:",attr"`
	TDsFromReturns                         StringInt `xml:",attr"`
	TimeOfPossession                       GameTime  `xml:",attr"`
	TotalExtraPointAttempts                StringInt `xml:",attr"`
	TotalExtraPointSuccesses               StringInt `xml:",attr"`
	TotalFirstDowns                        StringInt `xml:",attr"`
	TotalPlays                             StringInt `xml:",attr"`
	TotalReturnYardageNotIncludingKickoffs StringInt `xml:",attr"`
	TotalScore                             StringInt `xml:",attr"`
	TotalTouchdowns                        StringInt `xml:",attr"`
	TotalYards                             StringInt `xml:",attr"`
	TouchdownsAllOther                     StringInt `xml:",attr"`
	TouchdownsFumbleReturns                StringInt `xml:",attr"`
	TouchdownsInterceptionReturns          StringInt `xml:",attr"`
	TouchdownsKickoffReturns               StringInt `xml:",attr"`
	TouchdownsNonOffense                   StringInt `xml:",attr"`
	TouchdownsPuntReturns                  StringInt `xml:",attr"`
	Turnovers                              StringInt `xml:",attr"`
	TwoPointAttemptsPassing                StringInt `xml:",attr"`
	TwoPointAttemptsRushing                StringInt `xml:",attr"`
	TwoPointSuccessesPassing               StringInt `xml:",attr"`
	TwoPointSuccessesReturns               StringInt `xml:",attr"`
	TwoPointSuccessesRushing               StringInt `xml:",attr"`
}

type StatFileCumeStatHeader struct {
	Quarter         string    `xml:",attr"`
	Phase           string    `xml:",attr"`
	Season          StringInt `xml:",attr"`
	SeasonType      string    `xml:",attr"`
	Week            StringInt `xml:",attr"`
	Game_Date       string    `xml:",attr"`
	HomeClubCode    string    `xml:",attr"`
	VisitorClubCode string    `xml:",attr"`
	GameKey         StringInt `xml:",attr"`
	FileNumber      StringInt `xml:",attr"`
}

// https://www.nflgsis.com/gsis/documentation/Partners/StatIDs.html
const (
	StatIDRushingYardsMinus                 = 1
	StatIDPuntBlocked                       = 2
	StatIDFirstDownRushing                  = 3
	StatIDFirstDownPassing                  = 4
	StatIDFirstDownPenalty                  = 5
	StatIDThirdDownAttemptConverted         = 6
	StatIDThirdDownAttemptFailed            = 7
	StatIDFourthDownAttemptConverted        = 8
	StatIDFourthDownAttemptFailed           = 9
	StatIDRushingYards                      = 10
	StatIDRushingYardsTD                    = 11
	StatIDRushingYardsNoRush                = 12
	StatIDRushingYardsTDNoRush              = 13
	StatIDPassIncomplete                    = 14
	StatIDPassingYards                      = 15
	StatIDPassingYardsTD                    = 16
	StatIDPassingYardsNoPass                = 17
	StatIDPassingYardsTDNoPass              = 18
	StatIDInterceptionPasser                = 19
	StatIDSackYards                         = 20
	StatIDPassReceptionYards                = 21
	StatIDPassReceptionYardsTD              = 22
	StatIDPassReceptionYardsNoReception     = 23
	StatIDPassReceptionYardsTDNoReception   = 24
	StatIDInterceptionYards                 = 25
	StatIDInterceptionYardsTD               = 26
	StatIDInterceptionYardsNoInterception   = 27
	StatIDInterceptionYardsTDNoInterception = 28
	StatIDPuntingYards                      = 29
	StatIDPuntInside20                      = 30
	StatIDPuntIntoEndZone                   = 31
	StatIDPuntWithTouchback                 = 32
	StatIDPuntReturnYards                   = 33
	StatIDPuntReturnYardsTD                 = 34
	StatIDPuntReturnYardsNoReturn           = 35
	StatIDPuntReturnYardsTDNoReturn         = 36
	StatIDPuntOutOfBounds                   = 37
	StatIDPuntDownedNoReturn                = 38
	StatIDPuntFairCatch                     = 39
	StatIDPuntTouchbackNoReturn             = 40
	StatIDKickoffYards                      = 41
	StatIDKickoffInside20                   = 42
	StatIDKickoffIntoEndZone                = 43
	StatIDKickoffWithTouchback              = 44
	StatIDKickoffReturnYards                = 45
	StatIDKickoffReturnYardsTD              = 46
	StatIDKickoffReturnYardsNoReturn        = 47
	StatIDKickoffReturnYardsTDNoReturn      = 48
	StatIDKickoffOutOfBounds                = 49
	StatIDKickoffFairCatch                  = 50
	StatIDKickoffTouchback                  = 51
	StatIDFumbleForced                      = 52
	StatIDFumbleNotForced                   = 53
	StatIDFumbleOutOfBounds                 = 54
	StatIDOwnRecoveryYards                  = 55
	StatIDOwnRecoveryYardsTD                = 56
	StatIDOwnRecoveryYardsNoRecovery        = 57
	StatIDOwnRecoveryYardsTDNoRecovery      = 58
	StatIDOpponentRecoveryYards             = 59
	StatIDOpponentRecoveryYardsTD           = 60
	StatIDOpponentRecoveryYardsNoRecovery   = 61
	StatIDOpponentRecoveryYardsTDNoRecovery = 62
	StatIDMiscellaneousYards                = 63
	StatIDMiscellaneousYardsTD              = 64
	StatIDTimeout                           = 68
	StatIDFieldGoalMissedYards              = 69
	StatIDFieldGoalYards                    = 70
	StatIDFieldGoalBlockedOffense           = 71
	StatIDExtraPointGood                    = 72
	StatIDExtraPointFailed                  = 73
	StatIDExtraPointBlocked                 = 74
	StatID2PointRushGood                    = 75
	StatID2PointRushFailed                  = 76
	StatID2PointPassGood                    = 77
	StatID2PointPassFailed                  = 78
	StatIDSoloTackle                        = 79
	StatIDAssistedTackle                    = 80
	StatIDHalfTackle                        = 81
	StatIDTackleAssist                      = 82
	StatIDSackYardsDefense                  = 83
	StatIDHalfSackYardsDefense              = 84
	StatIDPassDefensed                      = 85
	StatIDPuntBlockedDefense                = 86
	StatIDExtraPointBlockedDefense          = 87
	StatIDFieldGoalBlockedDefense           = 88
	StatIDSafetyDefense                     = 89
	StatIDHalfSafetyDefense                 = 90
	StatIDForcedFumble                      = 91
	StatIDPenalty                           = 93
	StatIDTackledForLoss                    = 95
	StatIDExtraPointSafety                  = 96
	StatID2PointRushSafety                  = 99
	StatID2PointPassSafety                  = 100
	StatIDKickoffKickDowned                 = 102
	StatIDSackYardsNoSack                   = 103
	StatID2PointPassReceptionGood           = 104
	StatID2PointPassReceptionFailed         = 105
	StatIDFumbleLost                        = 106
	StatIDOwnKickoffRecovery                = 107
	StatIDOwnKickoffRecoveryTD              = 108
	StatIDQuarterbackHit                    = 110
	StatIDPassLengthCompletion              = 111
	StatIDPassLengthNoCompletion            = 112
	StatIDYardageGainedAfterCatch           = 113
	StatIDPassTarget                        = 115
	StatIDTackleForLoss                     = 120
	StatIDLongFieldGoalYards                = 201
	StatIDExtraPointDeuce                   = 211
	StatID2PointRushDeuce                   = 212
	StatID2PointPassDeuce                   = 213
	StatIDExtraPointAborted                 = 301
	StatIDHalfTackleForLoss                 = 401
	StatIDTackleForLossYardage              = 402
	StatIDDefensive2PointAttempts           = 403
	StatIDDefensive2PointConversions        = 404
	StatIDDefensiveExtraPointAttempts       = 405
	StatIDDefensiveExtraPointConversions    = 406
	StatIDKickoffLength                     = 410
	StatID2PointReturnGood                  = 420
)

type StatFilePlayStat struct {
	PlayID        StringInt `xml:",attr"`
	StatID        StringInt `xml:",attr"`
	ClubCode      string    `xml:",attr"`
	PlayerID      string    `xml:",attr"`
	PlayerName    string    `xml:",attr"`
	Yards         StatYards `xml:",attr"`
	UniformNumber string    `xml:",attr"`
}

type StatYards struct {
	Value *int
}

func (y *StatYards) Int() int {
	if y.Value == nil {
		return 0
	}
	return *y.Value
}

func (y StatYards) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	value := ""
	if y.Value != nil {
		value = strconv.Itoa(*y.Value)
	}
	return xml.Attr{
		Name:  name,
		Value: value,
	}, nil
}

func (y *StatYards) unmarshal(s string) error {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	y.Value = &n
	if err != nil {
		return fmt.Errorf("error unmarshaling yards: %w", err)
	}
	return nil
}

func (y *StatYards) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case int:
		y.Value = &v
		return nil
	case string:
		return y.unmarshal(v)
	default:
		return fmt.Errorf("unexpected string int type: %T", v)
	}
}

func (y *StatYards) UnmarshalXMLAttr(attr xml.Attr) error {
	return y.unmarshal(attr.Value)
}

type YardLine struct {
	team   *string
	number *int
	raw    string
}

func (l *YardLine) Team() *string {
	return l.team
}

func (l *YardLine) Number() int {
	return *l.number
}

func (l *YardLine) String() string {
	if l.number != nil {
		if l.team != nil {
			return fmt.Sprintf("%v %v", *(l.team), *(l.number))
		}
		return fmt.Sprintf("%v", *(l.number))
	}
	return ""
}

func (l YardLine) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: l.raw,
	}, nil
}

func (l *YardLine) unmarshal(s string) error {
	if s == "" {
		*l = YardLine{}
		return nil
	}
	values := strings.Split(s, " ")
	if len(values) > 2 {
		return fmt.Errorf("expected <= 2 values, found %d", len(values))
	}
	newYardLine := YardLine{
		raw: s,
	}
	if len(values) > 1 {
		newYardLine.team = &values[0]
	}
	if v, err := strconv.Atoi(values[len(values)-1]); err != nil {
		return err
	} else {
		newYardLine.number = &v
	}
	*l = newYardLine
	return nil
}

func (l *YardLine) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return l.unmarshal(s)
}

func (l *YardLine) UnmarshalXMLAttr(attr xml.Attr) error {
	return l.unmarshal(attr.Value)
}

const (
	PlayTypeGame              = 1
	PlayTypePlayFromScrimmage = 2
	PlayTypeTimeout           = 4
	PlayTypeTry               = 26
	PlayTypeFreeKick          = 32
	PlayTypeEndQuarter        = 42
	PlayTypeComment           = 60
	PlayTypeEndGame           = 66
)

type StatFileFieldGoals struct {
	VisitorFGAttempts StringInt `xml:",attr"`
	VisitorFGMade     StringInt `xml:",attr"`
	HomeFGAttempts    StringInt `xml:",attr"`
	HomeFGMade        StringInt `xml:",attr"`
}

type StatFilePunts struct {
	VisitorPunts     StringInt `xml:",attr"`
	VisitorPuntYards StringInt `xml:",attr"`
	HomePunts        StringInt `xml:",attr"`
	HomePuntYards    StringInt `xml:",attr"`
}

type StatFilePlay struct {
	ClockTime                        GameTime    `xml:",attr"`
	EndQuarterPlay                   StringInt   `xml:",attr"`
	Down                             StringInt   `xml:",attr"`
	Quarter                          StringInt   `xml:",attr"`
	YardsToGo                        StringInt   `xml:",attr"`
	PlayClock                        StringInt   `xml:",attr"`
	PlayID                           StringInt   `xml:",attr"`
	PlaySeq                          StringFloat `xml:",attr"`
	PlayType                         StringInt   `xml:",attr"`
	PlayDescription                  string      `xml:",attr"`
	PlayDescriptionWithJerseyNumbers string      `xml:",attr"`
	TimeOfDay                        string      `xml:",attr"`
	IsScoringPlay                    StringBool  `xml:",attr"`
	YardLine                         YardLine    `xml:",attr"`
	PossessionTeam                   string      `xml:",attr"`

	// Non-zero (not necessarily 1) when the play is deleted.
	PlayDeleted StringInt `xml:",attr"`

	// Time remaining in the quarter at the end of the play
	EndClockTime GameTime `xml:",attr"`
}

type StatFileScoringSummaryEvent struct {
	Quarter         StringInt `xml:",attr"`
	ClockTime       GameTime  `xml:",attr"`
	PlayDescription string    `xml:",attr"`
	VisitorScore    StringInt `xml:",attr"`
	HomeScore       StringInt `xml:",attr"`
	ScoreType       string    `xml:",attr"`
	ScoringPlayID   StringInt `xml:",attr"`
	PATPlayID       StringInt `xml:",attr"`
	ScoringClubCode string    `xml:",attr"`
}

// The plays in the stat file that haven't been deleted and aren't random things like commentary or
// metadata.
func (f *StatFile) ActualPlays() []*StatFilePlay {
	ret := make([]*StatFilePlay, 0, len(f.Play))
	for _, p := range f.Play {
		if p.IsActualPlay() {
			ret = append(ret, p)
		}
	}
	return ret
}

// Returns true if the play hasn't been deleted and isn't some random thing like commentary or
// metadata.
func (p *StatFilePlay) IsActualPlay() bool {
	if p.PlayDeleted != 0 {
		return false
	}
	switch p.PlayType {
	case PlayTypeGame, PlayTypeEndQuarter, PlayTypeComment, PlayTypeEndGame:
		return false
	}
	return true
}

type GameTime struct {
	minutes *int
	seconds *int
	raw     string
}

func (t *GameTime) IsNil() bool {
	return t.minutes == nil || t.seconds == nil
}

func (t *GameTime) Duration() time.Duration {
	if t.IsNil() {
		return 0
	}
	return time.Minute*time.Duration(*t.minutes) + time.Second*time.Duration(*t.seconds)
}

func (t GameTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: t.raw,
	}, nil
}

func (t *GameTime) unmarshal(s string) error {
	if s == "" {
		*t = GameTime{}
		return nil
	}
	values := strings.Split(s, ":")
	var minutes, seconds string
	switch len(values) {
	case 1: // MMSS
		secondsDigits := 2
		if len(s) < 2 {
			secondsDigits = len(s)
		}
		minutes, seconds = s[:len(s)-secondsDigits], s[len(s)-secondsDigits:]
	case 2: // MM:SS
		minutes, seconds = values[0], values[1]
	default:
		return fmt.Errorf("unsupported clock time string: %s", s)
	}
	newTime := GameTime{
		raw: s,
	}
	if minutes == "" {
		zero := 0
		newTime.minutes = &zero
	} else if v, err := strconv.Atoi(minutes); err != nil {
		return fmt.Errorf("error unmarshaling game time minutes: %w", err)
	} else {
		newTime.minutes = &v
	}
	if v, err := strconv.Atoi(seconds); err != nil {
		return fmt.Errorf("error unmarshaling game time seconds: %w", err)
	} else {
		newTime.seconds = &v
	}

	*t = newTime
	return nil
}

func (t *GameTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return t.unmarshal(s)
}

func (t *GameTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return t.unmarshal(attr.Value)
}

func ParseTimeOfDay(timeOfDay string, referenceTime time.Time) (time.Time, error) {
	parts := strings.Split(timeOfDay, ":")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid time of day: " + timeOfDay)
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	s, _ := strconv.Atoi(parts[2])
	d := 0
	utcReferenceTime := referenceTime.In(time.UTC)
	if h+12 < utcReferenceTime.Hour() {
		d += 1
	} else if utcReferenceTime.Hour()+12 < h {
		d -= 1
	}
	return time.Date(utcReferenceTime.Year(), utcReferenceTime.Month(), utcReferenceTime.Day()+d, h, m, s, 0, time.UTC), nil
}
