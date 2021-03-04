package gsis

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gsisIntegration = flag.Bool("gsis-integration", false, "if given, test gsis integration")

func TestClientIntegration(t *testing.T) {
	if !*gsisIntegration {
		t.Skipf("to run this test, pass the --gsis-integration flag")
	}

	c := &Client{}
	signalr := c.OpenSignalRClient(logrus.StandardLogger())
	t.Cleanup(func() {
		assert.NoError(t, signalr.Close())
	})

	t.Run("GetStatFile", func(t *testing.T) {
		statFile, err := signalr.GetStatFile(2019, "REG", 15, 58120)
		require.NoError(t, err)
		assert.True(t, len(statFile.Play) > 5)
	})

	t.Run("GetIncrementalStatFileXML", func(t *testing.T) {
		statFile, number, statFileTime, err := c.GetIncrementalStatFile(20191229, "SEA", 100)
		require.NoError(t, err)
		assert.Len(t, statFile.Play, 1)
		assert.Equal(t, 100, number)
		assert.Equal(t, time.Date(2019, time.December, 30, 2, 23, 23, 0, time.UTC), statFileTime)
	})

	t.Run("GetRosterFile", func(t *testing.T) {
		rosterFile, err := c.GetRosterFile(2019, "REG", 15, 58120)
		require.NoError(t, err)
		assert.True(t, len(rosterFile.Player) > 5)
	})

	t.Run("GetTeamLogoSVG", func(t *testing.T) {
		svg, err := c.GetTeamLogoSVG("ARZ")
		require.NoError(t, err)
		assert.NotEmpty(t, svg)
	})
}

func TestClient_GetCurrentWeek(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/CurrentWeek")
	}))
	defer s.Close()

	c := &Client{URL: s.URL}

	w, err := c.GetCurrentWeek()
	require.NoError(t, err)

	assert.Equal(t, &CurrentWeek{
		Season:     2020,
		SeasonType: "Post",
		Week:       4,
	}, w)
}

func TestClient_GetSchedule(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/2019-SF-SEA/2019/REG/17/Schedule")
	}))
	defer s.Close()

	c := &Client{URL: s.URL}

	schedule, err := c.GetSchedule(2019, "REG", 17)
	require.NoError(t, err)
	require.Len(t, schedule, 16)
	assert.Equal(t, &ScheduleGame{
		GameKey:         58155,
		GameDate:        "12/29/2019",
		HomeClubCode:    "SEA",
		VisitorClubCode: "SF",
	}, schedule[15])
}

func TestClient_GetPlayFeedJSON(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The API returns a JSON dict encoded as a JSON string. -_-
		w.Write([]byte(`"{\"playFeed\":{\"gameKey\":58199,\"situation\":{\"homeClub\":\"LV\",\"visitClub\":\"NO\",\"homeScore\":0,\"visitScore\":0,\"phase\":\"P\",\"down\":0,\"yardsToGo\":0,\"yardLine\":\"\",\"possession\":\"\",\"playReview\":false,\"clock\":[\"54:01\",\"25\"]},\"openPlays\":[{\"playID\":1,\"sequence\":1.0,\"quarter\":1,\"playType\":\"Game\",\"playSubType\":\"NULL\",\"playEnded\":false,\"events\":[{\"eventID\":1,\"code\":\"GA\",\"name\":\"Game\",\"message\":\"\",\"eventAttributes\":[{\"eventAttributeID\":150,\"value\":\"Jon Gruden\",\"valueDescription\":\"Jon Gruden\",\"type\":\"Char\",\"sysCode\":\"Home Head Coach\",\"name\":\"Home Head Coach\"},{\"eventAttributeID\":151,\"value\":\"Sean Payton\",\"valueDescription\":\"Sean Payton\",\"type\":\"Char\",\"sysCode\":\"Visitor Head Coach\",\"name\":\"Visitor Head Coach\"},{\"eventAttributeID\":149,\"value\":\"Allegiant Stadium\",\"valueDescription\":\"Allegiant Stadium\",\"type\":\"Char\",\"sysCode\":\"Stadium\",\"name\":\"Stadium\"},{\"eventAttributeID\":152,\"value\":\"Las Vegas, Nevada\",\"valueDescription\":\"Las Vegas, Nevada\",\"type\":\"Char\",\"sysCode\":\"Location\",\"name\":\"Location\"},{\"eventAttributeID\":618,\"value\":\"701\",\"valueDescription\":\"Closed Stadium\",\"type\":\"Lookup\",\"sysCode\":\"Stadium Type\",\"name\":\"Stadium Type\"},{\"eventAttributeID\":154,\"value\":\"Natural Grass\",\"valueDescription\":\"Natural Grass\",\"type\":\"Char\",\"sysCode\":\"Turf Type\",\"name\":\"Turf Type\"},{\"eventAttributeID\":155,\"value\":\"Pacific\",\"valueDescription\":\"Pacific\",\"type\":\"Char\",\"sysCode\":\"Time Zone\",\"name\":\"Time Zone\"},{\"eventAttributeID\":156,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"GameTime\",\"sysCode\":\"GameStart\",\"name\":\"Game Start Time\"},{\"eventAttributeID\":145,\"value\":\"Sunny\",\"valueDescription\":\"Sunny\",\"type\":\"Char\",\"sysCode\":\"Game Weather\",\"name\":\"Game Weather\"},{\"eventAttributeID\":22,\"value\":\"99\",\"valueDescription\":\"99\",\"type\":\"Number\",\"sysCode\":\"Temperature\",\"name\":\"Temperature\"},{\"eventAttributeID\":204,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Number\",\"sysCode\":\"Humidity\",\"name\":\"Humidity\"},{\"eventAttributeID\":143,\"value\":\"0\",\"valueDescription\":\"0\",\"type\":\"Char\",\"sysCode\":\"Wind Speed\",\"name\":\"Wind Speed\"},{\"eventAttributeID\":146,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Char\",\"sysCode\":\"Wind Direction\",\"name\":\"Wind Direction\"},{\"eventAttributeID\":147,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Char\",\"sysCode\":\"Wind Chill\",\"name\":\"Wind Chill\"},{\"eventAttributeID\":148,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Char\",\"sysCode\":\"Outdoor Weather\",\"name\":\"Outdoor Weather\"},{\"eventAttributeID\":501,\"value\":\"501\",\"valueDescription\":\"N/A (Indoors)\",\"type\":\"Lookup\",\"sysCode\":\"\",\"name\":\"Official Weather\"},{\"eventAttributeID\":159,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Team\",\"sysCode\":\"Club Won Coin Toss\",\"name\":\"Club Won Coin Toss\"},{\"eventAttributeID\":160,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Lookup\",\"sysCode\":\"Elects\",\"name\":\"Elects\"},{\"eventAttributeID\":161,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Lookup\",\"sysCode\":\"Other Club Elects\",\"name\":\"Other Club Elects\"},{\"eventAttributeID\":177,\"value\":\"122\",\"valueDescription\":\"Home\",\"type\":\"Lookup\",\"sysCode\":\"Game Home/Neutral\",\"name\":\"Game Home/Neutral\"},{\"eventAttributeID\":162,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Number\",\"sysCode\":\"Paid Attendance\",\"name\":\"Paid Attendance\"},{\"eventAttributeID\":195,\"value\":\"Hochuli, Shawn (83)\",\"valueDescription\":\"Hochuli, Shawn (83)\",\"type\":\"Official\",\"sysCode\":\"Referee\",\"name\":\"Referee\"},{\"eventAttributeID\":196,\"value\":\"George, Ramon (128)\",\"valueDescription\":\"George, Ramon (128)\",\"type\":\"Official\",\"sysCode\":\"Umpire\",\"name\":\"Umpire\"},{\"eventAttributeID\":197,\"value\":\"Thomas, Sarah (53)\",\"valueDescription\":\"Thomas, Sarah (53)\",\"type\":\"Official\",\"sysCode\":\"Down Judge\",\"name\":\"Down Judge\"},{\"eventAttributeID\":198,\"value\":\"Johnson, Carl (101)\",\"valueDescription\":\"Johnson, Carl (101)\",\"type\":\"Official\",\"sysCode\":\"Line Judge\",\"name\":\"Line Judge\"},{\"eventAttributeID\":199,\"value\":\"Dickson, Ryan (25)\",\"valueDescription\":\"Dickson, Ryan (25)\",\"type\":\"Official\",\"sysCode\":\"Field Judge\",\"name\":\"Field Judge\"},{\"eventAttributeID\":200,\"value\":\"Hill, Chad (125)\",\"valueDescription\":\"Hill, Chad (125)\",\"type\":\"Official\",\"sysCode\":\"Side Judge\",\"name\":\"Side Judge\"},{\"eventAttributeID\":201,\"value\":\"Martinez, Rich (39)\",\"valueDescription\":\"Martinez, Rich (39)\",\"type\":\"Official\",\"sysCode\":\"Back Judge\",\"name\":\"Back Judge\"},{\"eventAttributeID\":202,\"value\":\"Brown, Kevin (0)\",\"valueDescription\":\"Brown, Kevin (0)\",\"type\":\"Official\",\"sysCode\":\"Replay Official\",\"name\":\"Replay Official\"},{\"eventAttributeID\":232,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Char\",\"sysCode\":\"Game Summary Aux Title\",\"name\":\"Game Summary Aux Title\"},{\"eventAttributeID\":621,\"value\":\"\",\"valueDescription\":\"\",\"type\":\"Lookup\",\"sysCode\":\"Home Team Broadcast Side\",\"name\":\"Home Team Broadcast Side\"}]}]}],\"endedPlayIds\":[],\"scoreboard\":{\"gameClockTime\":\"54:01\",\"playClockTime\":\"25\",\"homeTeamName\":\"\",\"guestTeamName\":\"\",\"homeTeamScore\":0,\"guestTeamScore\":0,\"quarter\":1,\"ballOn\":0,\"down\":0,\"toGo\":0,\"homePossessionIndicator\":false,\"guestPossessionIndicator\":false,\"homeTimeoutsLeft\":3,\"guestTimeoutsLeft\":3},\"clockTransmitterVersion\":\"3.0.0.31354\",\"dateTimeStampUTC\":\"2020-09-21T23:20:59.2632267Z\",\"dateTimeStampOriginalUTC\":\"0001-01-01T00:00:00\",\"significantEvents\":{\"unofficialYardLine\":null,\"events\":[\"Bet Stop\"]},\"scoreboardHealthy\":true,\"statsHealthy\":true}}"`))
	}))
	defer s.Close()

	c := &Client{URL: s.URL}

	buf, err := c.GetPlayFeedJSON(58199, "")
	require.NoError(t, err)

	var data struct {
		GameKey int
	}
	require.NoError(t, json.Unmarshal(buf, &data))
	assert.Equal(t, 58199, data.GameKey)
}
