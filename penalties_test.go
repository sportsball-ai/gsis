package gsis

import (
	"encoding/json"
	"fmt"
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
		99: []PenaltyInfo{{
			FoulCode:        FoulCodeDefensiveHolding,
			Status:          PenaltyStatusAccepted,
			Team:            "DAL",
			Player:          PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "58"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("LA", 40),
		}},
		225: []PenaltyInfo{{
			FoulCode:        FoulCodeDelayOfGame,
			Status:          PenaltyStatusAccepted,
			Team:            "LA",
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("DAL", 41),
		}},
		474: []PenaltyInfo{{
			FoulCode:        FoulCodeNeutralZoneInfraction,
			Status:          PenaltyStatusAccepted,
			Team:            "DAL",
			Player:          PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "96"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("LA", 31),
		}},
		521: []PenaltyInfo{{
			FoulCode:        FoulCodeIllegalShift,
			Status:          PenaltyStatusAccepted,
			Team:            "LA",
			Player:          PenaltyPlayerInfo{Team: "LA", JerseyNumber: "83"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("LA", 49),
		}},
		799: []PenaltyInfo{
			{
				FoulCode:        FoulCodeDefensivePassInterference,
				Status:          PenaltyStatusOffsetting,
				Team:            "LA",
				Player:          PenaltyPlayerInfo{Team: "LA", JerseyNumber: "31"},
				EnforcementSpot: PenaltyEnforcementSpotPrevious,
				EnforcedAt:      NewYardLine("DAL", 34),
			},
			{
				FoulCode: FoulCodeOffensiveHolding,
				Status:   PenaltyStatusOffsetting,
				Team:     "DAL",
				Player:   PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "71"},
			},
		},
		1903: []PenaltyInfo{{
			FoulCode:        FoulCodeOffensiveHolding,
			Status:          PenaltyStatusAccepted,
			Team:            "LA",
			Player:          PenaltyPlayerInfo{Team: "LA", JerseyNumber: "77"},
			Distance:        10,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("LA", 25),
		}},
		2048: []PenaltyInfo{{
			FoulCode:        FoulCodeIllegalUseOfHands,
			Status:          PenaltyStatusAccepted,
			Team:            "LA",
			Player:          PenaltyPlayerInfo{Team: "LA", JerseyNumber: "90"},
			Distance:        4,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("LA", 7),
		}},
		3635: []PenaltyInfo{{
			FoulCode:        FoulCodeFalseStart,
			Status:          PenaltyStatusAccepted,
			Team:            "LA",
			Player:          PenaltyPlayerInfo{Team: "LA", JerseyNumber: "66"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("DAL", 30),
		}},
		3717: []PenaltyInfo{{
			FoulCode:        FoulCodeDefensiveOffside,
			Status:          PenaltyStatusAccepted,
			Team:            "DAL",
			Player:          PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "79"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("DAL", 29),
		}},
		3853: []PenaltyInfo{{
			FoulCode:        FoulCodeDefensiveOffside,
			Status:          PenaltyStatusAccepted,
			Team:            "DAL",
			Player:          PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "79"},
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("DAL", 2),
		}},
	} {
		found := false
		for _, p := range stats.Play {
			if int(p.PlayID) == playID {
				assert.Equal(t, expectedPenalties, ParsePlayDescriptionPenalties(p.PlayDescriptionWithJerseyNumbers), fmt.Sprintf("PlayID = %v, Description = %#v", playID, p.PlayDescriptionWithJerseyNumbers))
				found = true
			}
		}
		assert.True(t, found)
	}

	for desc, expected := range map[string][]PenaltyInfo{
		"(5:07) (Shotgun) 4-D.Prescott pass short left intended for 10-T.Austin INTERCEPTED by 31-D.Williams [50-S.Ebukam] at DAL 46. 31-D.Williams ran ob at DAL 46 for no gain. Penalty on LA-31-D.Williams, Defensive Pass Interference, offsetting, enforced at DAL 34 - No Play. Penalty on DAL-71-L.Collins, Offensive Holding, offsetting.": []PenaltyInfo{{
			FoulCode:        FoulCodeDefensivePassInterference,
			Status:          PenaltyStatusOffsetting,
			Team:            "LA",
			Player:          PenaltyPlayerInfo{Team: "LA", JerseyNumber: "31"},
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("DAL", 34),
		}, {
			FoulCode: FoulCodeOffensiveHolding,
			Status:   PenaltyStatusOffsetting,
			Team:     "DAL",
			Player:   PenaltyPlayerInfo{Team: "DAL", JerseyNumber: "71"},
		}},
		"(9:58) (Punt formation) Penalty on NE-38-B.Bolden, False Start, declined.": []PenaltyInfo{{
			FoulCode: FoulCodeFalseStart,
			Status:   PenaltyStatusDeclined,
			Team:     "NE",
			Player:   PenaltyPlayerInfo{Team: "NE", JerseyNumber: "38"},
		}},
		"(10:48) 12-T.Brady pass incomplete short middle to 16-J.Meyers [99-S.McLendon]. PENALTY on NYJ-34-B.Poole, Defensive Holding, 2 yards, enforced at NYJ 3 - No Play.": []PenaltyInfo{{
			FoulCode:        FoulCodeDefensiveHolding,
			Status:          PenaltyStatusAccepted,
			Team:            "NYJ",
			Player:          PenaltyPlayerInfo{Team: "NYJ", JerseyNumber: "34"},
			Distance:        2,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("NYJ", 3),
		}},
		"4-A.Seibert kicks 61 yards from CLV 35 to NYJ 4. 25-T.Cannon to NYJ 24 for 20 yards (44-S.Takitaki; 29-S.Redwine). PENALTY on NYJ-91-B.Kaufusi, Illegal Double-Team Block, 12 yards, enforced at NYJ 24.": []PenaltyInfo{{
			FoulCode:        FoulCodeIllegalDoubleTeamBlock,
			Status:          PenaltyStatusAccepted,
			Team:            "NYJ",
			Player:          PenaltyPlayerInfo{Team: "NYJ", JerseyNumber: "91"},
			Distance:        12,
			EnforcementSpot: PenaltyEnforcementSpotOther,
			EnforcedAt:      NewYardLine("NYJ", 24),
		}},
		"(:10) (Shotgun) 8-L.Falk pass short right to 88-T.Montgomery to NYJ 39 for 2 yards (51-M.Wilson). PENALTY on CLV-51-M.Wilson, Lowering the Head to Initiate Contact, 15 yards, enforced at NYJ 39.": []PenaltyInfo{{
			FoulCode:        FoulCodeUseOfHelmet,
			Status:          PenaltyStatusAccepted,
			Team:            "CLV",
			Player:          PenaltyPlayerInfo{Team: "CLV", JerseyNumber: "51"},
			Distance:        15,
			EnforcementSpot: PenaltyEnforcementSpotOther,
			EnforcedAt:      NewYardLine("NYJ", 39),
		}},
		"(:16) (Shotgun) 15-P.Mahomes pass short middle to 87-T.Kelce to HST 31 for 11 yards (41-Z.Cunningham). PENALTY on HST-41-Z.Cunningham, Horse Collar Tackle, 15 yards, enforced at HST 31.": []PenaltyInfo{{
			FoulCode:        FoulCodeHorseCollar,
			Status:          PenaltyStatusAccepted,
			Team:            "HST",
			Player:          PenaltyPlayerInfo{Team: "HST", JerseyNumber: "41"},
			Distance:        15,
			EnforcementSpot: PenaltyEnforcementSpotOther,
			EnforcedAt:      NewYardLine("HST", 31),
		}},
		"(5:44) 2-B.Colquitt punts 50 yards to IND 7, Center-58-A.Cutting, downed by MIN-85-D.Chisena. PENALTY on MIN-85-D.Chisena, Illegal Touch Kick, 5 yards, enforced at MIN 43 - No Play.": []PenaltyInfo{{
			FoulCode:        FoulCodeIllegalTouchKick,
			Status:          PenaltyStatusAccepted,
			Team:            "MIN",
			Player:          PenaltyPlayerInfo{Team: "MIN", JerseyNumber: "85"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("MIN", 43),
		}},
		"(8:45) 6-J.Scott punts 39 yards to DET 0, Center-43-H.Bradley, penalty enforced. PENALTY on DET-39-J.Agnew, Unnecessary Roughness, 5 yards, enforced at DET 10.": []PenaltyInfo{{
			FoulCode:        FoulCodeUnnecessaryRoughness,
			Status:          PenaltyStatusAccepted,
			Team:            "DET",
			Player:          PenaltyPlayerInfo{Team: "DET", JerseyNumber: "39"},
			Distance:        5,
			EnforcementSpot: PenaltyEnforcementSpotOther,
			EnforcedAt:      NewYardLine("DET", 10),
		}},
		"(Kick formation) TWO-POINT CONVERSION ATTEMPT. 6-M.Wishnowsky rushes left end. ATTEMPT FAILS. PENALTY on NYG-59-L.Carter, Face Mask (15 Yards), 8 yards, enforced at NYG 15 - No Play. Bad snap on the hold. 6-M.Wishnowsky attempts to rush and the penalty occurs.": []PenaltyInfo{{
			FoulCode:        FoulCodeFacemask,
			Status:          PenaltyStatusAccepted,
			Team:            "NYG",
			Player:          PenaltyPlayerInfo{Team: "NYG", JerseyNumber: "59"},
			Distance:        8,
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("NYG", 15),
		}},
		"(2:33) (Shotgun) 17-J.Allen pass short left to 14-S.Diggs for 1 yard, TOUCHDOWN NULLIFIED by Penalty [50-S.Ebukam]. Penalty on BUF-66-B.Winters, Offensive Holding, offsetting, enforced at LA 1 - No Play. Penalty on LA-50-S.Ebukam, Roughing the Passer, offsetting.": []PenaltyInfo{{
			FoulCode:        FoulCodeOffensiveHolding,
			Status:          PenaltyStatusOffsetting,
			Team:            "BUF",
			Player:          PenaltyPlayerInfo{Team: "BUF", JerseyNumber: "66"},
			EnforcementSpot: PenaltyEnforcementSpotPrevious,
			EnforcedAt:      NewYardLine("LA", 1),
		}, {
			FoulCode: FoulCodeRoughingThePasser,
			Status:   PenaltyStatusOffsetting,
			Team:     "LA",
			Player:   PenaltyPlayerInfo{Team: "LA", JerseyNumber: "50"},
		}},
	} {
		assert.Equal(t, expected, ParsePlayDescriptionPenalties(desc), desc)
	}
}
