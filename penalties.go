package gsis

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type FoulCode string

const (
	FoulCodeChopBlock                  FoulCode = "CHB"
	FoulCodeIllegalSubstitution        FoulCode = "ILS"
	FoulCodeClipping                   FoulCode = "CLP"
	FoulCodeIllegalTouchKick           FoulCode = "ITK"
	FoulCodeDefensiveDelayOfGame       FoulCode = "DOD"
	FoulCodeIllegalTouchPass           FoulCode = "ITP"
	FoulCodeDefensiveHolding           FoulCode = "DH"
	FoulCodeIllegalUseOfHands          FoulCode = "ILH"
	FoulCodeDefensiveOffside           FoulCode = "DOF"
	FoulCodeIllegalWedge               FoulCode = "WED"
	FoulCodeDefensivePassInterference  FoulCode = "DPI"
	FoulCodeIneligibleDownfieldKick    FoulCode = "IDK"
	FoulCodeDefensiveTooManyMenOnField FoulCode = "DTM"
	FoulCodeIneligibleDownfieldPass    FoulCode = "IDP"
	FoulCodeDelayOfGame                FoulCode = "DOG"
	FoulCodeIntentionalGrounding       FoulCode = "ING"
	FoulCodeDelayOfKickoff             FoulCode = "DOK"
	FoulCodeInvalidFairCatchSignal     FoulCode = "IFC"
	FoulCodeDisqualification           FoulCode = "DSQ"
	FoulCodeKickCatchInterference      FoulCode = "KCI"
	FoulCodeEncroachment               FoulCode = "ENC"
	FoulCodeKickoffOutOfBounds         FoulCode = "KOB"
	FoulCodeFacemask                   FoulCode = "FMM"
	FoulCodeLeaping                    FoulCode = "LEA"
	FoulCodeFairCatchInterference      FoulCode = "FCI"
	FoulCodeLeverage                   FoulCode = "LEV"
	FoulCodeFalseStart                 FoulCode = "FST"
	FoulCodeLowBlock                   FoulCode = "LBL"
	FoulCodeHorseCollar                FoulCode = "HC"
	FoulCodeNeutralZoneInfraction      FoulCode = "NZI"
	FoulCodeIllegalBat                 FoulCode = "BAT"
	FoulCodeOffensiveHolding           FoulCode = "OH"
	FoulCodeIllegalBlindsideBlock      FoulCode = "BLI"
	FoulCodeOffensiveOffside           FoulCode = "OOF"
	FoulCodeIllegalBlockAboveTheWaist  FoulCode = "IBW"
	FoulCodeOffensivePassInterference  FoulCode = "OPI"
	FoulCodeIllegalContact             FoulCode = "ICT"
	FoulCodeOffensiveTooManyMenOnField FoulCode = "OTM"
	FoulCodeIllegalCrackback           FoulCode = "ICB"
	FoulCodeOffsideOnFreeKick          FoulCode = "OFK"
	FoulCodeIllegalCut                 FoulCode = "ICU"
	FoulCodePlayerOutOfBoundsonKick    FoulCode = "POK"
	FoulCodeIllegalDoubleTeamBlock     FoulCode = "IDT"
	FoulCodeRoughingTheKicker          FoulCode = "RRK"
	FoulCodeIllegalFormation           FoulCode = "ILF"
	FoulCodeRoughingThePasser          FoulCode = "RPS"
	FoulCodeIllegalForwardHandoff      FoulCode = "IFH"
	FoulCodeRunningIntoTheKicker       FoulCode = "RNK"
	FoulCodeIllegalForwardPass         FoulCode = "IFP"
	FoulCodeTaunting                   FoulCode = "TAU"
	FoulCodeIllegalKick                FoulCode = "KIK"
	FoulCodeTripping                   FoulCode = "TRP"
	FoulCodeIllegalMotion              FoulCode = "ILM"
	FoulCodeUnnecessaryRoughness       FoulCode = "UNR"
	FoulCodeIllegalPeelBack            FoulCode = "IPB"
	FoulCodeUnsportsmanlikeConduct     FoulCode = "UNS"
	FoulCodeIllegalShift               FoulCode = "ISH"
	FoulCodeUseOfHelmet                FoulCode = "UOH"
)

var FoulCodesByDescription = map[string]FoulCode{
	"chop block":                            FoulCode("CHB"),
	"illegal substitution":                  FoulCode("ILS"),
	"clipping":                              FoulCode("CLP"),
	"illegal touch kick":                    FoulCode("ITK"),
	"defensive delay of game":               FoulCode("DOD"),
	"illegal touch pass":                    FoulCode("ITP"),
	"defensive holding":                     FoulCode("DH"),
	"illegal use of hands":                  FoulCode("ILH"),
	"defensive offside":                     FoulCode("DOF"),
	"illegal wedge":                         FoulCode("WED"),
	"defensive pass interference":           FoulCode("DPI"),
	"ineligible downfield kick":             FoulCode("IDK"),
	"defensive too many men on field":       FoulCode("DTM"),
	"ineligible downfield pass":             FoulCode("IDP"),
	"delay of game":                         FoulCode("DOG"),
	"intentional grounding":                 FoulCode("ING"),
	"delay of kickoff":                      FoulCode("DOK"),
	"invalid fair catch signal":             FoulCode("IFC"),
	"disqualification":                      FoulCode("DSQ"),
	"kick catch interference":               FoulCode("KCI"),
	"encroachment":                          FoulCode("ENC"),
	"kickoff out of bounds":                 FoulCode("KOB"),
	"facemask":                              FoulCode("FMM"),
	"face mask (15 yards)":                  FoulCode("FMM"),
	"leaping":                               FoulCode("LEA"),
	"fair catch interference":               FoulCode("FCI"),
	"leverage":                              FoulCode("LEV"),
	"false start":                           FoulCode("FST"),
	"low block":                             FoulCode("LBL"),
	"horse collar":                          FoulCode("HC"),
	"horse collar tackle":                   FoulCode("HC"),
	"neutral zone infraction":               FoulCode("NZI"),
	"illegal bat":                           FoulCode("BAT"),
	"offensive holding":                     FoulCode("OH"),
	"illegal blindside block":               FoulCode("BLI"),
	"offensive offside":                     FoulCode("OOF"),
	"illegal block above the waist":         FoulCode("IBW"),
	"offensive pass interference":           FoulCode("OPI"),
	"illegal contact":                       FoulCode("ICT"),
	"offensive too many men on field":       FoulCode("OTM"),
	"illegal crackback":                     FoulCode("ICB"),
	"offside on free kick":                  FoulCode("OFK"),
	"illegal cut":                           FoulCode("ICU"),
	"player out of bounds on kick":          FoulCode("POK"),
	"illegal double team block":             FoulCode("IDT"),
	"roughing the kicker":                   FoulCode("RRK"),
	"illegal formation":                     FoulCode("ILF"),
	"roughing the passer":                   FoulCode("RPS"),
	"illegal forward handoff":               FoulCode("IFH"),
	"running into the kicker":               FoulCode("RNK"),
	"illegal forward pass":                  FoulCode("IFP"),
	"taunting":                              FoulCode("TAU"),
	"illegal kick/kicking loose ball":       FoulCode("KIK"),
	"tripping":                              FoulCode("TRP"),
	"illegal motion":                        FoulCode("ILM"),
	"unnecessary roughness":                 FoulCode("UNR"),
	"illegal peel back":                     FoulCode("IPB"),
	"unsportsmanlike conduct":               FoulCode("UNS"),
	"illegal shift":                         FoulCode("ISH"),
	"use of helmet":                         FoulCode("UOH"),
	"lowering the head to initiate contact": FoulCode("UOH"),
}

type OffsettingPenalty struct {
	FoulCode     FoulCode
	Team         string
	JerseyNumber string
}

type PenaltyEnforcementSpot int

const (
	PenaltyEnforcementSpotNone PenaltyEnforcementSpot = iota
	PenaltyEnforcementSpotPrevious
	PenaltyEnforcementSpotDeadBall
	PenaltyEnforcementSpotSucceeding
	PenaltyEnforcementSpotOther
)

type PenaltyInfo struct {
	FoulCode        FoulCode
	Status          PenaltyStatus
	Team            string
	JerseyNumber    string
	EnforcementSpot PenaltyEnforcementSpot
	EnforcedAt      *YardLine
	Distance        int
}

type PenaltyStatus string

const (
	PenaltyStatusAccepted   PenaltyStatus = "ACCEPTED"
	PenaltyStatusDeclined   PenaltyStatus = "DECLINED"
	PenaltyStatusOffsetting PenaltyStatus = "OFFSETTING"
)

func ParsePlayDescriptionPenalties(description string) []PenaltyInfo {
	// Added to remove GSIS description noise
	cleanedDescription := strings.ReplaceAll(description, "TOUCHDOWN NULLIFIED by Penalty [", "TOUCHDOWN NULLIFIED by [")
	parts := strings.SplitN(strings.ToLower(cleanedDescription), "penalty ", 2)
	if len(parts) == 1 {
		return nil
	}
	cleanedDescription = "penalty " + parts[1]

	var ret []PenaltyInfo

	newPenalty := PenaltyInfo{
		Status: PenaltyStatusAccepted,
	}

	splitDesc := regexp.MustCompile(`[,.]\s`).Split(cleanedDescription, -1)

	for idx, part := range splitDesc {
		// For each penalty on a play, the order will go
		// 1. "penalty on {player/team}"
		// 2. Penalty description ("false start")
		// 3. Yards OR offsetting OR declined ("5 yards", "offsetting", "declined")
		// 4. Enforcement spot ("enforced at la 7 - no play.")

		part = strings.TrimSuffix(part, ".")

		// Remove GSIS noise section
		if part == "penalty enforced" {
			continue
		}

		// Append existing penalty if not first penalty
		if strings.Contains(part, "penalty ") && idx > 0 {
			// Catch cases where it says "penalty enforced" before actual penalty (1 seen so far)
			if !(idx == 1 && splitDesc[0] == "penalty enforced") {
				ret = append(ret, newPenalty)
				newPenalty = PenaltyInfo{
					Status: PenaltyStatusAccepted,
				}
			}
		}

		// Add foul code info
		if code, ok := FoulCodesByDescription[strings.ReplaceAll(part, "-", " ")]; ok {
			newPenalty.FoulCode = code
		}

		// Add player/team info
		if strings.HasPrefix(part, "penalty on ") {
			subject := strings.TrimPrefix(part, "penalty on ")
			parts := strings.Split(subject, "-")
			teamAbbreviation := strings.ToUpper(parts[0])
			newPenalty.Team = teamAbbreviation
			if len(parts) > 1 {
				number := parts[1]
				if len(number) < 2 || !unicode.IsDigit(rune(number[1])) {
					number = "0" + number
				}
				newPenalty.JerseyNumber = number
			}
		}

		if part == "enforced between downs" {
			newPenalty.EnforcementSpot = PenaltyEnforcementSpotSucceeding
		}

		if strings.HasPrefix(part, "enforced at ") {
			yardLineString := strings.TrimPrefix(part, "enforced at ")
			yardLineString = strings.Split(yardLineString, " - ")[0]
			parts := strings.Split(yardLineString, " ")
			if number, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				l := &YardLine{
					number: &number,
				}
				if len(parts) > 1 {
					team := strings.ToUpper(parts[0])
					l.team = &team
				}
				newPenalty.EnforcedAt = l
				newPenalty.EnforcementSpot = PenaltyEnforcementSpotOther
			}
		}

		if strings.HasSuffix(part, " yards") {
			yardsString := strings.TrimSuffix(part, " yards")
			if n, err := strconv.Atoi(yardsString); err == nil {
				newPenalty.Distance = n
			}
		}

		if strings.HasSuffix(part, "no play") {
			newPenalty.EnforcementSpot = PenaltyEnforcementSpotPrevious
		}

		// Update status if necessary
		if strings.Contains(part, "declined") {
			newPenalty.Status = PenaltyStatusDeclined
		} else if strings.Contains(part, "offsetting") {
			newPenalty.Status = PenaltyStatusOffsetting
		}
	}

	// Append last penalty of play if valid (in case GSIS adds random penalty text somewhere)
	if newPenalty.FoulCode != "" && newPenalty.Team != "" {
		ret = append(ret, newPenalty)
	}
	return ret
}
