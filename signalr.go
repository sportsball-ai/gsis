package gsis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type SignalRClient struct {
	// A SignalR URL, such as "https://www.nflgsis.com/GameStatsLive/signalr".
	URL string

	// For example: `[{"name":"gamestatshub"},{"name":"schedulehub"}]`.
	ConnectionData string

	Logger logrus.FieldLogger

	conn             *SignalRConnection
	connectError     error
	connectErrorTime time.Time
	connMutex        sync.Mutex
}

type SignalRConnection struct {
	conn   *websocket.Conn
	logger logrus.FieldLogger

	outgoing chan *websocket.PreparedMessage

	readLoopDone     chan struct{}
	writeLoopDone    chan struct{}
	beginClosingOnce sync.Once
	close            chan struct{}

	invocationMutex    sync.Mutex
	invocationChannels map[int]chan *SignalRServerMessage
	invocationsClosed  bool
	nextInvocationId   int
}

func NewSignalRConnection(conn *websocket.Conn, logger logrus.FieldLogger) *SignalRConnection {
	ret := &SignalRConnection{
		conn:               conn,
		logger:             logger,
		outgoing:           make(chan *websocket.PreparedMessage, 100),
		readLoopDone:       make(chan struct{}),
		writeLoopDone:      make(chan struct{}),
		close:              make(chan struct{}),
		invocationChannels: make(map[int]chan *SignalRServerMessage),
		nextInvocationId:   0,
	}
	go ret.readLoop()
	go ret.writeLoop()
	return ret
}

func (c *SignalRConnection) readLoop() {
	defer close(c.readLoopDone)
	defer c.beginClosing()

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				select {
				case <-c.close:
				default:
					c.logger.Error(fmt.Errorf("websocket read error: %w", err))
				}
			}
			return
		}
		c.handleMessage(p)
	}
}

func (c *SignalRConnection) writeLoop() {
	defer c.finishClosing()
	defer close(c.writeLoopDone)
	defer close(c.outgoing)

	defer c.conn.Close()

	for {
		var msg *websocket.PreparedMessage
		select {
		case outgoing, ok := <-c.outgoing:
			if !ok {
				return
			}
			msg = outgoing
		case <-c.close:
			return
		}

		c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		if err := c.conn.WritePreparedMessage(msg); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway, websocket.CloseNormalClosure) && err != websocket.ErrCloseSent {
				c.logger.Error(fmt.Errorf("websocket write error: %w", err))
			}
			return
		}
	}
}

func (c *SignalRConnection) handleMessage(buf []byte) {
	var msg SignalRServerMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		c.logger.Error("error unmarshaling server message: %w", err)
		return
	}

	var ch chan *SignalRServerMessage

	if msg.I != "" {
		if id, err := strconv.Atoi(msg.I); err == nil {
			c.invocationMutex.Lock()
			ch = c.invocationChannels[id]
			c.invocationMutex.Unlock()
		}
	}

	if ch != nil {
		select {
		case ch <- &msg:
		default:
		}
	}
}

func (c *SignalRConnection) beginClosing() {
	c.beginClosingOnce.Do(func() {
		close(c.close)
	})
}

func (c *SignalRConnection) finishClosing() {
	<-c.readLoopDone
	<-c.writeLoopDone
	c.invocationMutex.Lock()
	defer c.invocationMutex.Unlock()
	c.invocationsClosed = true
	for _, ch := range c.invocationChannels {
		close(ch)
	}
	c.invocationChannels = nil
}

func (c *SignalRConnection) Close() error {
	c.beginClosing()
	c.finishClosing()
	return nil
}

func (c *SignalRConnection) IsClosed() bool {
	select {
	case <-c.close:
		return true
	default:
		return false
	}
}

type SignalRClientMessage struct {
	// The hub.
	H string

	// The name of the method.
	M string

	// The method arguments.
	A []interface{}

	// An invocation number used to identify the response.
	I int
}

type SignalRServerMessage struct {
	// The invocation number that the server is responding to.
	I string

	// The payload if this is a response to a client message.
	R json.RawMessage
}

var ErrSignalRConnectionClosed = fmt.Errorf("signalr connection closed")

func (c *SignalRConnection) Invoke(ctx context.Context, hub, method string, args ...interface{}) (json.RawMessage, error) {
	c.invocationMutex.Lock()
	if c.invocationsClosed {
		c.invocationMutex.Unlock()
		return nil, ErrSignalRConnectionClosed
	}
	ch := make(chan *SignalRServerMessage, 1)
	id := c.nextInvocationId
	for {
		if _, ok := c.invocationChannels[id]; !ok {
			c.invocationChannels[id] = ch
			break
		}
		id = (id + 1) % 0x80000000
		if id == c.nextInvocationId {
			return nil, fmt.Errorf("no unallocated invocation ids")
		}
	}
	c.nextInvocationId = (id + 1) % 0x80000000
	c.invocationMutex.Unlock()

	defer func() {
		c.invocationMutex.Lock()
		delete(c.invocationChannels, id)
		c.invocationMutex.Unlock()
	}()

	msg := &SignalRClientMessage{
		H: hub,
		M: method,
		A: args,
		I: id,
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error serializing client message: %w", err)
	}

	p, err := websocket.NewPreparedMessage(websocket.TextMessage, buf)
	if err != nil {
		return nil, fmt.Errorf("error preparing client message: %w", err)
	}

	select {
	case c.outgoing <- p:
	default:
		return nil, fmt.Errorf("output buffer full")
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, ErrSignalRConnectionClosed
		}
		return resp.R, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *SignalRClient) Connection() (*SignalRConnection, error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	// We have a connection.
	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn, nil
	}

	// We have a recent connect error.
	now := time.Now()
	if c.connectError != nil && now.Sub(c.connectErrorTime) < time.Second {
		return nil, c.connectError
	}

	// No connection or recent error. Try to connect.
	if conn, err := c.connect(); err != nil {
		c.connectError = err
		c.connectErrorTime = now
	} else {
		c.conn = NewSignalRConnection(conn, c.Logger)
		c.connectError = nil
	}
	return c.conn, c.connectError
}

func (c *SignalRClient) Close() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn.Close()
	}
	return nil
}

func (c *SignalRClient) connect() (*websocket.Conn, error) {
	for attempt := 0; ; attempt++ {
		resp, err := c.doNegotiateRequest()
		if err != nil {
			if attempt >= 2 {
				return nil, fmt.Errorf("negotiate error: %w", err)
			}
			time.Sleep(time.Second)
			continue
		}
		if resp.ConnectionToken == "" {
			return nil, fmt.Errorf("no connection token in negotiate response")
		} else if resp.URL == "" {
			return nil, fmt.Errorf("no url in negotiate response")
		} else if !resp.TryWebSockets {
			return nil, fmt.Errorf("server does not support websockets")
		}

		signalrURL, err := url.Parse(c.URL + "/")
		if err != nil {
			return nil, fmt.Errorf("error parsing url: %w", err)
		}

		connectURL := signalrURL.ResolveReference(&url.URL{
			Path: "connect",
			RawQuery: url.Values{
				"transport":       []string{"webSockets"},
				"clientProtocol":  []string{"1.5"},
				"connectionToken": []string{resp.ConnectionToken},
			}.Encode(),
		})
		if connectURL.Scheme == "http" {
			connectURL.Scheme = "ws"
		} else {
			connectURL.Scheme = "wss"
		}

		conn, _, err := websocket.DefaultDialer.Dial(connectURL.String(), nil)
		if err != nil {
			if attempt >= 2 {
				return nil, fmt.Errorf("websocket dial error: %w", err)
			}
			time.Sleep(time.Second)
			continue
		}
		return conn, nil
	}
}

type SignalRNegotiateResponse struct {
	URL             string
	ConnectionToken string
	TryWebSockets   bool
}

func (c *SignalRClient) doNegotiateRequest() (*SignalRNegotiateResponse, error) {
	signalrURL, err := url.Parse(c.URL + "/")
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}

	negotiateURL := signalrURL.ResolveReference(&url.URL{
		Path: "negotiate",
		RawQuery: url.Values{
			"clientProtocol": []string{"1.5"},
			"connectionData": []string{c.ConnectionData},
		}.Encode(),
	})

	resp, err := http.Get(negotiateURL.String())
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	var result SignalRNegotiateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &result, nil
}

func (c *SignalRClient) GetStatFile(year int, season string, week, gameKey int) (*StatFile, error) {
	buf, err := c.GetStatFileJSON(year, season, week, gameKey)
	if err != nil {
		return nil, err
	}
	var statFile StatFile
	if err := json.Unmarshal(buf, &statFile); err != nil {
		return nil, fmt.Errorf("error unmarshaling stat file: %w", err)
	}
	return &statFile, nil
}

func (c *SignalRClient) GetStatFileJSON(year int, season string, week, gameKey int) (json.RawMessage, error) {
	conn, err := c.Connection()
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := conn.Invoke(ctx, "schedulehub", "RegisterForSchedule", strconv.Itoa(year), strings.ToUpper(season), week); err != nil {
		return nil, fmt.Errorf("error registering for schedule: %w", err)
	}

	defer func() {
		if _, err := conn.Invoke(ctx, "gamestatshub", "UnregisterForStats", ""); err != nil {
			c.Logger.Warn(fmt.Errorf("error unregistering for stats: %w", err))
		}
	}()

	if statsJSON, err := conn.Invoke(ctx, "gamestatshub", "RegisterForStats", strconv.Itoa(gameKey)); err != nil {
		return nil, fmt.Errorf("error registering for stats: %w", err)
	} else {
		return statsJSON, nil
	}
}
