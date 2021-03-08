package gsis

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	// The GSIS URL. By default this is "https://www.nflgsis.com".
	URL string

	// The URL of the entry domain. By default this is the URL with the "https://www." replaced by
	// "https://entry.".
	EntryURL string

	// The URL of the services domain. By default this is the URL with the "https://www." replaced by
	// "https://services.".
	ServicesURL string
}

func (c *Client) url() string {
	if c.URL != "" {
		return c.URL
	}
	return "https://www.nflgsis.com"
}

func (c *Client) entryURL() string {
	if c.EntryURL != "" {
		return c.EntryURL
	} else if url := c.url(); strings.HasPrefix(url, "https://www.") {
		return "https://entry." + strings.TrimPrefix(url, "https://www.")
	} else {
		return url
	}
}

func (c *Client) servicesURL() string {
	if c.ServicesURL != "" {
		return c.ServicesURL
	} else if url := c.url(); strings.HasPrefix(url, "https://www.") {
		return "https://services." + strings.TrimPrefix(url, "https://www.")
	} else {
		return url
	}
}

type CurrentWeek struct {
	Season     int
	SeasonType string
	Week       int
}

func (c *Client) GetCurrentWeek() (*CurrentWeek, error) {
	resp, err := http.Get(strings.TrimSuffix(c.url(), "/") + "/CurrentWeek")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	ret := &CurrentWeek{}
	parts := bytes.Split(buf, []byte{0xb8})
	if len(parts) < 3 {
		return nil, fmt.Errorf("not enough response fields")
	} else if ret.Season, err = strconv.Atoi(string(parts[0])); err != nil {
		return nil, fmt.Errorf("error parsing season: %w", err)
	} else if ret.Week, err = strconv.Atoi(string(parts[2])); err != nil {
		return nil, fmt.Errorf("error parsing week: %w", err)
	}
	ret.SeasonType = strings.TrimSpace(string(parts[1]))
	return ret, nil
}

type ScheduleGame struct {
	// M/D/Y
	GameDate        string
	GameKey         int
	HomeClubCode    string
	VisitorClubCode string
}

func (c *Client) GetSchedule(season int, seasonType string, week int) ([]*ScheduleGame, error) {
	resp, err := http.Get(strings.TrimSuffix(c.url(), "/") + fmt.Sprintf("/%d/%v/%02d/Schedule", season, seasonType, week))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	ret := []*ScheduleGame{}
	for _, line := range bytes.Split(buf, []byte{0x0a}) {
		parts := bytes.Split(line, []byte{0xb8})
		if len(parts) < 17 {
			continue
		}
		game := &ScheduleGame{
			GameDate:        string(parts[1]),
			HomeClubCode:    string(parts[4]),
			VisitorClubCode: string(parts[11]),
		}
		if game.GameKey, err = strconv.Atoi(string(parts[0])); err != nil {
			return nil, fmt.Errorf("error parsing game key: %w", err)
		}
		ret = append(ret, game)
	}
	return ret, nil
}

func (c *Client) GetCumulativeStatFile(date int, homeClubCode string) (*StatFile, int, time.Time, error) {
	buf, number, t, err := c.GetCumulativeStatFileXML(date, homeClubCode)
	if err != nil {
		return nil, 0, t, err
	}
	var statFile StatFile
	if err := xml.Unmarshal(buf, &statFile); err != nil {
		return nil, 0, t, fmt.Errorf("error unmarshaling cumulative stat file: %w", err)
	}
	return &statFile, number, t, nil
}

func (c *Client) GetCumulativeStatFileXML(date int, homeClubCode string) ([]byte, int, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf(strings.TrimSuffix(c.entryURL(), "/")+"/DataInterfaceServer/%v/%v/gametodate", date, strings.ToUpper(homeClubCode))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, time.Time{}, fmt.Errorf("error creating cumulative stat file request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, time.Time{}, fmt.Errorf("error getting cumulative stat file xml: %w", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, 0, time.Time{}, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, 0, time.Time{}, fmt.Errorf("unexpected cumulative stat file status code: %v", resp.StatusCode)
	}

	t, err := time.Parse("20060102 150405", resp.Header.Get("gsisfiletimestamp"))
	if err != nil {
		return nil, 0, time.Time{}, fmt.Errorf("error parsing gsis file timestamp: %w", err)
	}

	number, _ := strconv.Atoi(resp.Header.Get("gsisfilenumber"))

	buf, err := ioutil.ReadAll(resp.Body)
	return buf, number, t, err
}

func (c *Client) GetIncrementalStatFile(date int, homeClubCode string, number int) (*StatFile, int, time.Time, error) {
	buf, number, t, err := c.GetIncrementalStatFileXML(date, homeClubCode, number)
	if err != nil {
		return nil, 0, t, err
	}
	var statFile StatFile
	if err := xml.Unmarshal(buf, &statFile); err != nil {
		return nil, 0, t, fmt.Errorf("error unmarshaling incremental stat file: %w", err)
	}
	return &statFile, number, t, nil
}

// Gets an incremental stat file, returning immediately if it is unavailable.
func (c *Client) GetIncrementalStatFileXML(date int, homeClubCode string, number int) ([]byte, int, time.Time, error) {
	return c.LongPollIncrementalStatFileXML(date, homeClubCode, number, 0)
}

// Gets an incremental stat file, blocking until it is available.
func (c *Client) LongPollIncrementalStatFileXML(date int, homeClubCode string, number int, timeoutSeconds int) ([]byte, int, time.Time, error) {
	return c.longPollIncrementalXML("StatXML", date, homeClubCode, number, timeoutSeconds)
}

// Gets an incremental roster file, returning immediately if it is unavailable.
func (c *Client) GetIncrementalRosterFileXML(date int, homeClubCode string, number int) ([]byte, int, time.Time, error) {
	return c.LongPollIncrementalRosterFileXML(date, homeClubCode, number, 0)
}

// Gets an incremental roster file, blocking until it is available.
func (c *Client) LongPollIncrementalRosterFileXML(date int, homeClubCode string, number int, timeoutSeconds int) ([]byte, int, time.Time, error) {
	return c.longPollIncrementalXML("RosterXML", date, homeClubCode, number, timeoutSeconds)
}

func (c *Client) longPollIncrementalXML(name string, date int, homeClubCode string, number int, timeoutSeconds int) ([]byte, int, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds+10)*time.Second)
	defer cancel()

	url := fmt.Sprintf(strings.TrimSuffix(c.entryURL(), "/")+"/DataInterfaceServer/%v/%v/%v/%v?timeout=%d", date, strings.ToUpper(homeClubCode), name, number, timeoutSeconds)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, time.Time{}, fmt.Errorf("error creating incremental stat file request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, time.Time{}, fmt.Errorf("error getting incremental stat file xml: %w", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, 0, time.Time{}, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, 0, time.Time{}, fmt.Errorf("unexpected incremental stat file status code: %v", resp.StatusCode)
	}

	t, err := time.Parse("20060102 150405", resp.Header.Get("gsisfiletimestamp"))
	if err != nil {
		return nil, 0, time.Time{}, fmt.Errorf("error parsing gsis file timestamp: %w", err)
	}

	if number == 0 {
		number, _ = strconv.Atoi(resp.Header.Get("gsisfilenumber"))
	}

	buf, err := ioutil.ReadAll(resp.Body)
	return buf, number, t, err
}

func (c *Client) GetRosterFile(year int, season string, week, gameKey int) (*RosterFile, error) {
	buf, err := c.GetRosterFileXML(year, season, week, gameKey)
	if err != nil {
		return nil, err
	}
	var rosterFile RosterFile
	if err := xml.Unmarshal(buf, &rosterFile); err != nil {
		return nil, fmt.Errorf("error unmarshaling roster stat file: %w", err)
	}
	return &rosterFile, nil
}

func (c *Client) GetRosterFileXML(year int, season string, week, gameKey int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(strings.TrimSuffix(c.url(), "/")+"/%v/%v/%02d/%v/Roster.xml", year, season, week, gameKey), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating roster request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting roster: %w", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected roster status code: %v", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) GetTeamLogoSVG(clubCode string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(strings.TrimSuffix(c.url(), "/")+"/GameStatsLive/Images/SVG_Knockout/NFL/%v.svg", strings.ToUpper(clubCode)), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating logo svg request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting logo svg: %w", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected logo svg status code: %v", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) GetPlayFeedJSON(gameKey int, token string) (json.RawMessage, error) {
	req, err := http.NewRequest("GET", strings.TrimSuffix(c.servicesURL(), "/")+"/GSISClockSituation/PlayFeed/"+strconv.Itoa(gameKey), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	// for some reason the response json is encoded into a json string
	var body string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}
	var data struct {
		PlayFeed json.RawMessage
	}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}
	return data.PlayFeed, nil
}

func (c *Client) OpenSignalRClient(logger logrus.FieldLogger) *SignalRClient {
	return &SignalRClient{
		URL:            strings.TrimSuffix(c.url(), "/") + "/GameStatsLive/signalr",
		ConnectionData: `[{"name":"gamestatshub"},{"name":"schedulehub"}]`,
		Logger:         logger,
	}
}
