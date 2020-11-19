package gsis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalRClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/negotiate" {
			w.Write([]byte(`{"Url":"/","ConnectionToken":"token","TryWebSockets":true}`))
		} else if r.URL.Path == "/connect" {
			conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
			require.NoError(t, err)

			messageType, p, err := conn.ReadMessage()
			require.NoError(t, err)
			assert.Equal(t, websocket.TextMessage, messageType)
			require.JSONEq(t, `{"H": "schedulehub", "M": "RegisterForSchedule", "A":["2019","REG",3], "I":0}`, string(p))

			assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"R": {"foo":"bar"}, "I": "0"}`)))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := &SignalRClient{
		URL:            ts.URL,
		ConnectionData: `[{"name":"gamestatshub"},{"name":"schedulehub"}]`,
	}

	conn, err := c.Connection()
	require.NoError(t, err)

	resp, err := conn.Invoke(context.Background(), "schedulehub", "RegisterForSchedule", "2019", "REG", 3)
	require.NoError(t, err)
	assert.JSONEq(t, `{"foo":"bar"}`, string(resp))
}
