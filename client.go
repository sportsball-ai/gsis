package gsis

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

func (c *Client) GetIncrementalStatFile(date int, homeClubCode string, number int) (*StatFile, time.Time, error) {
	buf, t, err := c.GetIncrementalStatFileXML(date, homeClubCode, number)
	if err != nil {
		return nil, t, err
	}
	var statFile StatFile
	if err := xml.Unmarshal(buf, &statFile); err != nil {
		return nil, t, fmt.Errorf("error unmarshaling incremental stat file: %w", err)
	}
	return &statFile, t, nil
}

func (c *Client) GetIncrementalStatFileXML(date int, homeClubCode string, number int) ([]byte, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(strings.TrimSuffix(c.entryURL(), "/")+"/DataInterfaceServer/%v/%v/STATXML/%v", date, strings.ToUpper(homeClubCode), number), nil)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("error creating incremental stat file request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("error getting incremental stat file xml: %w", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, time.Time{}, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, time.Time{}, fmt.Errorf("unexpected incremental stat file status code: %v", resp.StatusCode)
	}

	t, err := time.Parse("20060102 150405", resp.Header.Get("gsisfiletimestamp"))
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("error parsing gsis file timestamp: %w", err)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	return buf, t, err
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

func (c *Client) OpenSignalRClient(logger logrus.FieldLogger) *SignalRClient {
	return &SignalRClient{
		URL:            strings.TrimSuffix(c.url(), "/") + "/GameStatsLive/signalr",
		ConnectionData: `[{"name":"gamestatshub"},{"name":"schedulehub"}]`,
		Logger:         logger,
	}
}
