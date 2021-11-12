package janus

import . "github.com/jinuopti/lilpop-server/log"

// Handle represents a handle to a plugin instance on the Gateway.
type Handle struct {
	// ID is the handle_id of this plugin handle
	ID uint64

	// Type   // pub  or sub
	Type string

	//User   // Userid
	User string

	// Events is a receive only channel that can be used to receive events
	// related to this handle from the gateway.
	Events chan interface{}

	session *Session
}

func (handle *Handle) send(msg map[string]interface{}, transaction chan interface{}) {
	msg["handle_id"] = handle.ID
	handle.session.send(msg, transaction)
}

// Request sends a sync request
func (handle *Handle) Request(body interface{}) (*SuccessMsg, error) {
	req, ch := newRequest("message")
	if body != nil {
		req["body"] = body
	}
	handle.send(req, ch)

	msg := <-ch

	switch msg := msg.(type) {
	case *SuccessMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("message")
}

// Message sends a message request to a plugin handle on the Gateway.
// body should be the plugin data to be passed to the plugin, and jsep should
// contain an optional SDP offer/answer to establish a WebRTC PeerConnection.
// On success, an EventMsg will be returned and error will be nil.
func (handle *Handle) Message(body, jsep interface{}) (*EventMsg, error) {
	req, ch := newRequest("message")
	if body != nil {
		req["body"] = body
	}
	if jsep != nil {
		req["jsep"] = jsep
	}
	Logd("Message: %s", req)
	handle.send(req, ch)

GetMessage: // No tears..
	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		goto GetMessage // ..only dreams.
	case *EventMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("message")
}

func (handle *Handle) MessageCreate(body interface{}) (*SuccessMsg, error) {
	req, ch := newRequest("message")
	if body != nil {
		req["body"] = body
	}
	Logd("Message: %s", req)
	handle.send(req, ch)

GetMessage: // No tears..
	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		goto GetMessage // ..only dreams.
	case *SuccessMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("message")
}

// Trickle sends a trickle request to the Gateway as part of establishing
// a new PeerConnection with a plugin.
// candidate should be a single ICE candidate, or a completed object to
// signify that all candidates have been sent:
//		{
//			"completed": true
//		}
// On success, an AckMsg will be returned and error will be nil.
func (handle *Handle) Trickle(candidate interface{}) (*AckMsg, error) {
	req, ch := newRequest("trickle")
	req["candidate"] = candidate
	handle.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("trickle")
}

// TrickleMany sends a trickle request to the Gateway as part of establishing
// a new PeerConnection with a plugin.
// candidates should be an array of ICE candidates.
// On success, an AckMsg will be returned and error will be nil.
func (handle *Handle) TrickleMany(candidates interface{}) (*AckMsg, error) {
	req, ch := newRequest("trickle")
	req["candidates"] = candidates
	handle.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("trickle")
}

// Detach sends a detach request to the Gateway to remove this handle.
// On success, an AckMsg will be returned and error will be nil.
func (handle *Handle) Detach() (*AckMsg, error) {
	req, ch := newRequest("detach")
	handle.send(req, ch)

	var ack *AckMsg
	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		ack = msg
	case *ErrorMsg:
		return nil, msg
	}

	// Remove this handle from the session
	handle.session.Lock()
	delete(handle.session.Handles, handle.ID)
	handle.session.Unlock()

	return ack, nil
}
