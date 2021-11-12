package janus

import (
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/jinuopti/lilpop-server/log"
	"github.com/rs/xid"
)

var (
	Client *Gateway
)

func newRequest(method string) (map[string]interface{}, chan interface{}) {
	req := make(map[string]interface{}, 8)
	req["janus"] = method
	return req, make(chan interface{})
}

func generateTransactionId() xid.ID {
	return xid.New()
}

// Connect initiates a webscoket connection with the Janus Gateway
func Connect(address string, port string) (*Gateway, error) {
	if Client != nil {
		return Client, nil
	}

	websocket.DefaultDialer.Subprotocols = []string{"janus-protocol"}

	url := "ws://" + address + ":" + port
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	Client = new(Gateway)
	Client.conn = conn
	Client.transactions = make(map[xid.ID]chan interface{})
	Client.transactionsUsed = make(map[xid.ID]bool)
	Client.Sessions = make(map[uint64]*Session)
	Client.sendChan = make(chan []byte, 100)
	Client.errors = make(chan error)

	go Client.ping()
	go Client.recv()

	Logd("New Janus WebSocket client connect")

	return Client, nil
}

func passMsg(ch chan interface{}, msg interface{}) {
	ch <- msg
}

func unexpected(request string) error {
	return fmt.Errorf("Unexpected response received to '%s' request", request)
}
