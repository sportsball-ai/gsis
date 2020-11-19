package gsis

import (
	"flag"
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
		statFile, statFileTime, err := c.GetIncrementalStatFile(20191229, "SEA", 100)
		require.NoError(t, err)
		assert.Len(t, statFile.Play, 1)
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
