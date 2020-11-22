package gsis

type RosterFile struct {
	GameKey *RosterFileGameKey
	Player  []RosterFilePlayer
}

type RosterFileGameKey struct {
	GameKey       int    `xml:",attr"`
	GameDate      string `xml:",attr"`
	HomeClubKey   int    `xml:",attr"`
	VisitClubKey  int    `xml:",attr"`
	HomeClubName  string `xml:",attr"`
	VisitClubName string `xml:",attr"`
	HomeClubCode  string `xml:",attr"`
	VisitClubCode string `xml:",attr"`
}

type RosterFilePlayer struct {
	GSISPlayer_ID string `xml:",attr"`
	ClubKey       int    `xml:",attr"`
	Name          string `xml:",attr"`
	Status        string `xml:",attr"`
	JerseyNumber  string `xml:",attr"`
	Position      string `xml:",attr"`
	FirstName     string `xml:",attr"`
	LastName      string `xml:",attr"`
}
