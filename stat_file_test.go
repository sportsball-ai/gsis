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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatFile(t *testing.T) {
	f, err := os.Open("testdata/signalr-stats.json")
	require.NoError(t, err)
	defer f.Close()

	var stats StatFile
	require.NoError(t, json.NewDecoder(f).Decode(&stats))

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	files, err := ioutil.ReadDir("testdata/games")
	require.NoError(t, err)
	for _, f := range files {
		if f.IsDir() {
			t.Run(f.Name(), func(t *testing.T) {
				f, err := os.Open(filepath.Join("testdata/games", f.Name(), "GSISGameStats.xml"))
				require.NoError(t, err)
				defer f.Close()

				var stats StatFile
				require.NoError(t, xml.NewDecoder(f).Decode(&stats))
			})
		}
	}
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

	finalXML, err := ioutil.ReadFile("testdata/2019-SF-SEA/2019/REG/17/58155/GSISGameStats.xml")
	require.NoError(t, err)
	var finalStatFileFromXML *StatFile
	require.NoError(t, xml.Unmarshal(finalXML, &finalStatFileFromXML))

	finalJSON, err := ioutil.ReadFile("testdata/2019-SF-SEA/2019/REG/17/58155/signalr-stats.json")
	require.NoError(t, err)
	var finalStatFileFromJSON *StatFile
	require.NoError(t, json.Unmarshal(finalJSON, &finalStatFileFromJSON))
	// SignalR doesn't seem to emit the down judge correctly
	finalStatFileFromJSON.GameAttributes.DownJudge = "McKenzie, Dana (8)"
	// SignalR also doesn't do nullified stats correctly
	finalStatFileFromJSON.PlayStatNullified = statFile.PlayStatNullified

	for _, p := range statFile.Play {
		// the incremental stat XML files seem to include newlines where the other files just uses spaces
		p.PlayDescription = strings.ReplaceAll(p.PlayDescription, "\n", " ")
		p.PlayDescriptionWithJerseyNumbers = strings.ReplaceAll(p.PlayDescriptionWithJerseyNumbers, "\n", " ")
	}

	for name, expected := range map[string]*StatFile{
		"XML":  finalStatFileFromXML,
		"JSON": finalStatFileFromJSON,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, expected.CumeStatHeader, statFile.CumeStatHeader)
			assert.Equal(t, expected.Play, statFile.Play)
			assert.Equal(t, expected.PlayStat, statFile.PlayStat)
			assert.Equal(t, expected.PlayStatNullified, statFile.PlayStatNullified)
			assert.Equal(t, expected.HomeTeamStats, statFile.HomeTeamStats)
			assert.Equal(t, expected.GameAttributes, statFile.GameAttributes)
			assert.Equal(t, expected.VisitorTeamStats, statFile.VisitorTeamStats)
			assert.Equal(t, expected.ScoringSummary, statFile.ScoringSummary)

			assert.Equal(t, expected, statFile)
		})
	}
}

func TestGameTime(t *testing.T) {
	for input, expected := range map[string]time.Duration{
		"01:02": time.Minute + 2*time.Second,
		"0102":  time.Minute + 2*time.Second,
		"102":   time.Minute + 2*time.Second,
		"1:12":  time.Minute + 12*time.Second,
		":12":   12 * time.Second,
		"12":    12 * time.Second,
	} {
		buf, err := json.Marshal(input)
		require.NoError(t, err)
		var actual GameTime
		require.NoError(t, json.Unmarshal(buf, &actual), err)
		assert.Equal(t, expected, actual.Duration())
	}
}
