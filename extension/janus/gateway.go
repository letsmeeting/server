package janus

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	utility "github.com/jinuopti/lilpop-server/library"
	. "github.com/jinuopti/lilpop-server/log"
	"github.com/rs/xid"
	"log"
	"sync"
	"time"
)

// Gateway represents a connection to an instance of the Janus Gateway.
type Gateway struct {
	// Sessions is a map of the currently active sessions to the gateway.
	Sessions map[uint64]*Session

	// Access to the Sessions map should be synchronized with the Gateway.Lock()
	// and Gateway.Unlock() methods provided by the embeded sync.Mutex.
	sync.Mutex

	conn             *websocket.Conn
	transactions     map[xid.ID]chan interface{}
	transactionsUsed map[xid.ID]bool
	errors           chan error
	sendChan         chan []byte
	writeMu          sync.Mutex
}

// Close closes the underlying connection to the Gateway.
func (gateway *Gateway) Close() error {
	return gateway.conn.Close()
}

// GetErrChan returns a channels through which the caller can check and react to connectivity errors
func (gateway *Gateway) GetErrChan() chan error {
	return gateway.errors
}

func (gateway *Gateway) send(msg map[string]interface{}, transaction chan interface{}) {
	guid := generateTransactionId()

	// { "transaction": transaction_id }
	msg["transaction"] = guid.String()
	gateway.Lock()
	gateway.transactions[guid] = transaction
	gateway.transactionsUsed[guid] = false
	gateway.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		Logd("json.Marshal: %s", err)
		return
	}

	// Logging
	prettyJson, err := utility.GetPrettyJsonStr(data)
	if err == nil {
		Logd("Write: %s", prettyJson)
	}

	gateway.writeMu.Lock()
	err = gateway.conn.WriteMessage(websocket.TextMessage, data)
	gateway.writeMu.Unlock()

	if err != nil {
		select {
		case gateway.errors <- err:
		default:
			Logd("conn.Write: %s\n", err)
		}
		return
	}
}

func (gateway *Gateway) ping() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := gateway.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(20*time.Second))
			if err != nil {
				select {
				case gateway.errors <- err:
				default:
					log.Println("ping:", err)
				}

				return
			}
		}
	}
}

func (gateway *Gateway) sendloop() {

}

func (gateway *Gateway) recv() {

	for {
		// Read message from Gateway

		// Decode to Msg struct
		var base BaseMsg

		_, data, err := gateway.conn.ReadMessage()
		if err != nil {
			select {
			case gateway.errors <- err:
			default:
				fmt.Printf("conn.Read: %s\n", err)
			}

			return
		}

		if err := json.Unmarshal(data, &base); err != nil {
			fmt.Printf("json.Unmarshal: %s\n", err)
			continue
		}

		// log message being sent
		prettyJson, err := utility.GetPrettyJsonStr(data)
		if err == nil {
			Logd("Recv: -> \n%s", prettyJson)
		}

		typeFunc, ok := msgtypes[base.Type]
		if !ok {
			fmt.Printf("Unknown message type received!\n")
			continue
		}

		msg := typeFunc()
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Printf("json.Unmarshal: %s\n", err)
			continue // Decode error
		}

		var transactionUsed bool
		if base.ID != "" {
			id, _ := xid.FromString(base.ID)
			gateway.Lock()
			transactionUsed = gateway.transactionsUsed[id]
			gateway.Unlock()

		}

		// Pass message on from here
		if base.ID == "" || transactionUsed {
			// Is this a Handle event?
			if base.Handle == 0 {
				// Error()
			} else {
				// Lookup Session
				gateway.Lock()
				session := gateway.Sessions[base.Session]
				gateway.Unlock()
				if session == nil {
					fmt.Printf("Unable to deliver message. Session gone?\n")
					continue
				}

				// Lookup Handle
				session.Lock()
				handle := session.Handles[base.Handle]
				session.Unlock()
				if handle == nil {
					fmt.Printf("Unable to deliver message. Handle gone?\n")
					continue
				}

				// Pass msg
				go passMsg(handle.Events, msg)
			}
		} else {
			id, _ := xid.FromString(base.ID)
			// Lookup Transaction
			gateway.Lock()
			transaction := gateway.transactions[id]
			switch msg.(type) {
			case *EventMsg:
				gateway.transactionsUsed[id] = true
			}
			gateway.Unlock()
			if transaction == nil {
				// Error()
			}

			// Pass msg
			go passMsg(transaction, msg)
		}
	}
}

// Info sends an info request to the Gateway.
// On success, an InfoMsg will be returned and error will be nil.
func (gateway *Gateway) Info() (*InfoMsg, error) {
	req, ch := newRequest("info")
	gateway.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *InfoMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("info")
}

// Create sends a create request to the Gateway.
// On success, a new Session will be returned and error will be nil.
func (gateway *Gateway) Create() (*Session, error) {
	req, ch := newRequest("create")
	gateway.send(req, ch)

	msg := <-ch
	var success *SuccessMsg
	switch msg := msg.(type) {
	case *SuccessMsg:
		success = msg
	case *ErrorMsg:
		return nil, msg
	}

	// Create new session
	session := new(Session)
	session.gateway = gateway
	session.ID = success.Data.ID
	session.Handles = make(map[uint64]*Handle)
	session.Events = make(chan interface{}, 2)

	// Store this session
	gateway.Lock()
	gateway.Sessions[session.ID] = session
	gateway.Unlock()

	return session, nil
}

func (gateway *Gateway) GetInfo() interface{} {
	var err error
	mess, err := gateway.Info()
	if err != nil {
		Loge("error: %s", err)
		return ""
	}
	return mess
}
