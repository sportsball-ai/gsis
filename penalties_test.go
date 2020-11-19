package gsis

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePlayDescriptionPenalties(t *testing.T) {
	f, err := os.Open("testdata/signalr-stats.json")
	require.NoError(t, err)
	defer f.Close()

	var stats StatFile
	require.NoError(t, json.NewDecoder(f).Decode(&stats))

	for playID, expectedPenalties := range map[int][]PenaltyInfo{
		99:   []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDefensiveHolding, Status: PenaltyStatusAccepted, Team: "DAL", Player: PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "58"}}},
		225:  []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDelayOfGame, Status: PenaltyStatusAccepted, Team: "LA"}},
		474:  []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeNeutralZoneInfraction, Status: PenaltyStatusAccepted, Team: "DAL", Player: PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "96"}}},
		521:  []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeIllegalShift, Status: PenaltyStatusAccepted, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "83"}}},
		799:  []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDefensivePassInterference, Status: PenaltyStatusOffsetting, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "31"}}, PenaltyInfo{FoulCode: FoulCodeOffensiveHolding, Status: PenaltyStatusOffsetting, Team: "DAL", Player: PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "71"}}},
		1903: []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeOffensiveHolding, Status: PenaltyStatusAccepted, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "77"}}},
		2048: []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeIllegalUseOfHands, Status: PenaltyStatusAccepted, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "90"}}},
		3635: []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeFalseStart, Status: PenaltyStatusAccepted, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "66"}}},
		3717: []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDefensiveOffside, Status: PenaltyStatusAccepted, Team: "DAL", Player: PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "79"}}},
		3853: []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDefensiveOffside, Status: PenaltyStatusAccepted, Team: "DAL", Player: PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "79"}}},
	} {
		found := false
		for _, p := range stats.Play {
			if int(p.PlayID) == playID {
				assert.Equal(t, expectedPenalties, ParsePlayDescriptionPenalties(p.PlayDescriptionWithJerseyNumbers))
				found = true
			}
		}
		assert.True(t, found)
	}

	for desc, expected := range map[string][]PenaltyInfo{
		"(5:07) (Shotgun) 4-D.Prescott pass short left intended for 10-T.Austin INTERCEPTED by 31-D.Williams [50-S.Ebukam] at DAL 46. 31-D.Williams ran ob at DAL 46 for no gain. Penalty on LA-31-D.Williams, Defensive Pass Interference, offsetting, enforced at DAL 34 - No Play. Penalty on DAL-71-L.Collins, Offensive Holding, offsetting.": []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDefensivePassInterference, Status: PenaltyStatusOffsetting, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "31"}}, PenaltyInfo{FoulCode: FoulCodeOffensiveHolding, Status: PenaltyStatusOffsetting, Team: "DAL", Player: PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "71"}}},
		"(9:58) (Punt formation) Penalty on NE-38-B.Bolden, False Start, declined.":                                                                                                                                                                                               []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeFalseStart, Status: PenaltyStatusDeclined, Team: "NE", Player: PenaltyPlayerInfo{Team: "NE", JerseyNumber: "38"}}},
		"(10:48) 12-T.Brady pass incomplete short middle to 16-J.Meyers [99-S.McLendon]. PENALTY on NYJ-34-B.Poole, Defensive Holding, 2 yards, enforced at NYJ 3 - No Play.":                                                                                                     []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeDefensiveHolding, Status: PenaltyStatusAccepted, Team: "NYJ", Player: PenaltyPlayerInfo{Team: "NYJ", JerseyNumber: "34"}}},
		"4-A.Seibert kicks 61 yards from CLV 35 to NYJ 4. 25-T.Cannon to NYJ 24 for 20 yards (44-S.Takitaki; 29-S.Redwine). PENALTY on NYJ-91-B.Kaufusi, Illegal Double-Team Block, 12 yards, enforced at NYJ 24.":                                                                []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeIllegalDoubleTeamBlock, Status: PenaltyStatusAccepted, Team: "NYJ", Player: PenaltyPlayerInfo{Team: "NYJ", JerseyNumber: "91"}}},
		"(:10) (Shotgun) 8-L.Falk pass short right to 88-T.Montgomery to NYJ 39 for 2 yards (51-M.Wilson). PENALTY on CLV-51-M.Wilson, Lowering the Head to Initiate Contact, 15 yards, enforced at NYJ 39.":                                                                      []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeUseOfHelmet, Status: PenaltyStatusAccepted, Team: "CLV", Player: PenaltyPlayerInfo{Team: "CLV", JerseyNumber: "51"}}},
		"(:16) (Shotgun) 15-P.Mahomes pass short middle to 87-T.Kelce to HST 31 for 11 yards (41-Z.Cunningham). PENALTY on HST-41-Z.Cunningham, Horse Collar Tackle, 15 yards, enforced at HST 31.":                                                                               []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeHorseCollar, Status: PenaltyStatusAccepted, Team: "HST", Player: PenaltyPlayerInfo{Team: "HST", JerseyNumber: "41"}}},
		"(5:44) 2-B.Colquitt punts 50 yards to IND 7, Center-58-A.Cutting, downed by MIN-85-D.Chisena. PENALTY on MIN-85-D.Chisena, Illegal Touch Kick, 5 yards, enforced at MIN 43 - No Play.":                                                                                   []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeIllegalTouchKick, Status: PenaltyStatusAccepted, Team: "MIN", Player: PenaltyPlayerInfo{Team: "MIN", JerseyNumber: "85"}}},
		"(8:45) 6-J.Scott punts 39 yards to DET 0, Center-43-H.Bradley, penalty enforced. PENALTY on DET-39-J.Agnew, Unnecessary Roughness, 5 yards, enforced at DET 10.":                                                                                                         []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeUnnecessaryRoughness, Status: PenaltyStatusAccepted, Team: "DET", Player: PenaltyPlayerInfo{Team: "DET", JerseyNumber: "39"}}},
		"(Kick formation) TWO-POINT CONVERSION ATTEMPT. 6-M.Wishnowsky rushes left end. ATTEMPT FAILS. PENALTY on NYG-59-L.Carter, Face Mask (15 Yards), 8 yards, enforced at NYG 15 - No Play. Bad snap on the hold. 6-M.Wishnowsky attempts to rush and the penalty occurs.":    []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeFacemask, Status: PenaltyStatusAccepted, Team: "NYG", Player: PenaltyPlayerInfo{Team: "NYG", JerseyNumber: "59"}}},
		"(2:33) (Shotgun) 17-J.Allen pass short left to 14-S.Diggs for 1 yard, TOUCHDOWN NULLIFIED by Penalty [50-S.Ebukam]. Penalty on BUF-66-B.Winters, Offensive Holding, offsetting, enforced at LA 1 - No Play. Penalty on LA-50-S.Ebukam, Roughing the Passer, offsetting.": []PenaltyInfo{PenaltyInfo{FoulCode: FoulCodeOffensiveHolding, Status: PenaltyStatusOffsetting, Team: "BUF", Player: PenaltyPlayerInfo{Team: "BUF", JerseyNumber: "66"}}, PenaltyInfo{FoulCode: FoulCodeRoughingThePasser, Status: PenaltyStatusOffsetting, Team: "LA", Player: PenaltyPlayerInfo{Team: "LA", JerseyNumber: "50"}}},
	} {
		assert.Equal(t, expected, ParsePlayDescriptionPenalties(desc))
	}
}
