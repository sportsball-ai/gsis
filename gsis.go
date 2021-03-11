package gsis

var gsisClubCodeToCommonAbbreviation = map[string]string{
	"LA":  "LAR",
	"BLT": "BAL",
	"HST": "HOU",
	"CLV": "CLE",
	"ARZ": "ARI",
}

// Returns the team's abbreviation based on the club code found in the GSIS API. The abbreviation is
// the common one, which is not necessarily the official one. These follow most consistent rules:
//
// * Single-word markets use the first three letters of the market (e.g. "ARI" for Arizona).
// * Multi-word markets use an initialism of the market (e.g. "GB" for Green Bar).
// * Where necessary, the first letter of the team's nickname disambiguates (e.g. "NYG" and "NYJ").
// * The Jaguars are an exception and are "JAX".
func CommonTeamAbbreviation(clubCode string) string {
	if abbr, ok := gsisClubCodeToCommonAbbreviation[clubCode]; ok {
		return abbr
	}
	return clubCode
}
