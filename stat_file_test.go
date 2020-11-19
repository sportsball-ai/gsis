package gsis

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatFile(t *testing.T) {
	f, err := os.Open("testdata/signalr-stats.json")
	require.NoError(t, err)
	defer f.Close()

	var stats StatFile
	require.NoError(t, json.NewDecoder(f).Decode(&stats))
}

func TestStatFileOfficials(t *testing.T) {
	for input, expected := range map[StatFileOfficials][]StatFileOfficial{
		"": nil,
		"Nathan Jones (42)": []StatFileOfficial{
			{
				FirstName:    "Nathan",
				LastName:     "Jones",
				JerseyNumber: "42",
			},
		},
		"Coleman IV, Walt (65)": []StatFileOfficial{
			{
				FirstName:    "Walt",
				LastName:     "Coleman",
				JerseyNumber: "65",
			},
		},
		"Walt Coleman IV (65)": []StatFileOfficial{
			{
				FirstName:    "Walt",
				LastName:     "Coleman",
				JerseyNumber: "65",
			},
		},
		"Walker, Jabir (26)": []StatFileOfficial{
			{
				FirstName:    "Jabir",
				LastName:     "Walker",
				JerseyNumber: "26",
			},
		},
		"Frantz, Earnie & Valenti, Terri": []StatFileOfficial{
			{
				FirstName: "Earnie",
				LastName:  "Frantz",
			},
			{
				FirstName: "Terri",
				LastName:  "Valenti",
			},
		},
	} {
		assert.Equal(t, expected, input.Officials(), input)
	}
}

func TestStatFile_Update(t *testing.T) {
	statFile := &StatFile{}
	for i := 1; i <= 271; i++ {
		buf, err := ioutil.ReadFile(filepath.Join("testdata", "2019-SF-SEA", "DataInterfaceServer", "20191229", "SEA", "STATXML", strconv.Itoa(i)))
		require.NoError(t, err)
		var update StatFile
		require.NoError(t, xml.Unmarshal(buf, &update))
		statFile.Update(&update)
	}

	finalJSON, err := ioutil.ReadFile("testdata/2019-SF-SEA/2019/REG/17/58155/signalr-stats.json")
	require.NoError(t, err)
	var finalStatFile *StatFile
	require.NoError(t, json.Unmarshal(finalJSON, &finalStatFile))

	// SignalR doesn't seem to emit the down judge correctly
	statFile.GameAttributes.DownJudge = ""

	for _, p := range statFile.Play {
		// XML files seem to include newlines where SignalR JSON just uses spaces
		p.PlayDescription = strings.ReplaceAll(p.PlayDescription, "\n", " ")
		p.PlayDescriptionWithJerseyNumbers = strings.ReplaceAll(p.PlayDescriptionWithJerseyNumbers, "\n", " ")
	}

	assert.Equal(t, finalStatFile.CumeStatHeader, statFile.CumeStatHeader)
	assert.Equal(t, finalStatFile.Play, statFile.Play)
	assert.ElementsMatch(t, finalStatFile.PlayStat, statFile.PlayStat)
	assert.Equal(t, finalStatFile.HomeTeamStats, statFile.HomeTeamStats)
	assert.Equal(t, finalStatFile.GameAttributes, statFile.GameAttributes)
	assert.Equal(t, finalStatFile.VisitorTeamStats, statFile.VisitorTeamStats)
	assert.Equal(t, finalStatFile.ScoringSummary, statFile.ScoringSummary)

	assert.Equal(t, finalStatFile, statFile)
}
