package gsis

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRosterFile(t *testing.T) {
	f, err := os.Open("testdata/Roster.xml")
	require.NoError(t, err)
	defer f.Close()

	var roster RosterFile
	require.NoError(t, xml.NewDecoder(f).Decode(&roster))
}
