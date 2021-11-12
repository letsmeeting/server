// Msg Types
//
// All messages received from the gateway are first decoded to the BaseMsg
// type. The BaseMsg type extracts the following JSON from the message:
//		{
//			"janus": <Type>,
//			"transaction": <ID>,
//			"session_id": <Session>,
//			"sender": <Handle>
//		}
// The Type field is inspected to determine which concrete type
// to decode the message to, while the other fields (ID/Session/Handle) are
// inspected to determine where the message should be delivered. Messages
// with an ID field defined are considered responses to previous requests, and
// will be passed directly to requester. Messages without an ID field are
// considered unsolicited events from the gateway and are expected to have
// both Session and Handle fields defined. They will be passed to the Events
// channel of the related Handle and can be read from there.

package janus

var msgtypes = map[string]func() interface{}{
	"error":       func() interface{} { return &ErrorMsg{} },
	"success":     func() interface{} { return &SuccessMsg{} },
	"detached":    func() interface{} { return &DetachedMsg{} },
	"server_info": func() interface{} { return &InfoMsg{} },
	"ack":         func() interface{} { return &AckMsg{} },
	"event":       func() interface{} { return &EventMsg{} },
	"webrtcup":    func() interface{} { return &WebRTCUpMsg{} },
	"media":       func() interface{} { return &MediaMsg{} },
	"hangup":      func() interface{} { return &HangupMsg{} },
	"slowlink":    func() interface{} { return &SlowLinkMsg{} },
	"timeout":     func() interface{} { return &TimeoutMsg{} },
	"join":		   func() interface{} { return &JoinMsg{} },
}

type BaseMsg struct {
	Type    string `json:"janus"`
	ID      string `json:"transaction"`
	Session uint64 `json:"session_id"`
	Handle  uint64 `json:"sender"`
}

type ErrorMsg struct {
	Err ErrorData `json:"error"`
}

type ErrorData struct {
	Code   int
	Reason string
}

func (err *ErrorMsg) Error() string {
	return err.Err.Reason
}

type SuccessMsg struct {
	Data       SuccessData
	PluginData PluginData
	Session    uint64 `json:"session_id"`
	Handle     uint64 `json:"sender"`
}

type SuccessData struct {
	ID uint64
}

type DetachedMsg struct{}

type InfoMsg struct {
	Name          string
	Version       int
	VersionString string `json:"version_string"`
	Author        string
	DataChannels  bool   `json:"data_channels"`
	IPv6          bool   `json:"ipv6"`
	LocalIP       string `json:"local-ip"`
	IceTCP        bool   `json:"ice-tcp"`
	Transports    map[string]PluginInfo
	Plugins       map[string]PluginInfo
}

type PluginInfo struct {
	Name          string
	Author        string
	Description   string
	Version       int
	VersionString string `json:"version_string"`
}

type AckMsg struct{}

type EventMsg struct {
	Plugindata PluginData
	Jsep       map[string]interface{}
	Session    uint64 `json:"session_id"`
	Handle     uint64 `json:"sender"`
}

type PluginData struct {
	Plugin string
	Data   map[string]interface{}
}

type WebRTCUpMsg struct {
	Session uint64 `json:"session_id"`
	Handle  uint64 `json:"sender"`
}

type TimeoutMsg struct {
	Session uint64 `json:"session_id"`
}

type SlowLinkMsg struct {
	Uplink bool
	Lost   int64
}

type MediaMsg struct {
	Type      string
	Receiving bool
}

type HangupMsg struct {
	Reason  string
	Session uint64 `json:"session_id"`
	Handle  uint64 `json:"sender"`
}

// JoinMsg Request message join room
type JoinMsg struct {
	Request 	string	`json:"request"` 	// "join"
	Room		float64	`json:"room"` 		// 1234
	Ptype		string	`json:"ptype"` 		// "publisher" "subscriber"
	Display		string	`json:"display"`	// "myname"
}

type OfferBody struct {
	Request 	string	`json:"request"`
	Audio 		bool	`json:"audio"`
	Video 		bool	`json:"video"`
}
type OfferJsep struct {
	Type 	string		`json:"type"`
	Sdp 	string 		`json:"sdp"`
}

type AnswerBody struct {
	Request     string  `json:"request"`
	Room        float64 `json:"room"`
}

type PluginVideoData struct {
	VideoRoom 	string	`json:"videoroom" example:"event"`
	Room 		float64 `json:"room" example:"1234"`
	Configured  string 	`json:"configured" example:"ok"`
	AudioCodec	string 	`json:"audio_codec" example:"opus"`
	VideoCodec  string 	`json:"video_codec" example:"vp8"`
}

type PluginVideoDataStarted struct {
	VideoRoom 	string	`json:"videoroom" example:"event"`
	Room 		float64 `json:"room" example:"1234"`
	Started     string  `json:"started" example:"ok"`
}

type Jsep struct {
	Type 	string	`json:"type" example:"answer"`
	Sdp 	string 	`json:"sdp"`
}

type PublisherInfo struct {
	Id 			float64		`json:"id"`
	Display 	string		`json:"display"`
	AudioCodec  string		`json:"audio_codec" example:"opus"`
	VideoCodec 	string		`json:"video_codec" example:"vp8"`
	Talking 	bool 		`json:"talking" example:"false"`
}

type VideoPluginDataEventAttached struct {
	Plugin    string 	`json:"plugin"`
	Display   string 	`json:"display"`
	Id        float64 	`json:"id"`
	Room      float64    `json:"room"`
	Videoroom string 	`json:"videoroom"`
}

type JoinSubscribe struct {
	Ptype     string 	`json:"ptype"`
	Request   string 	`json:"request"`
	Feed      float64 	`json:"feed"`
	PrivateId float64	`json:"private_id"`
	Room      float64  	`json:"room"`
}

type RoomSimpleReq struct {
	Request         string      `json:"request"`    // "list"
}

type RoomCreateReq struct {
	Request         string      `json:"request"`    // "create"
	Room            float64     `json:"room"`       // <unique numeric ID, optional, chosen by plugin if missing>
	Publishers      float64     `json:"publishers"` // <max number of concurrent senders>
	Description     string      `json:"description"`// <pretty name of the room, optional>
	VideoCodec      string      `json:"videocodec"` // vp8|vp9|h264|av1|h265
	BitRate         float64     `json:"bitrate"`    // <max video bitrate for senders> (e.g., 128000)
	Record          bool        `json:"record"`
	RecDir         *string      `json:"red_dir,omitempty"`
}

type RoomRsp struct {
	VideoRoom       string      `json:"videoroom"`  // "created|edited|destroyed"
	Room           *float64     `json:"room,omitempty"`
	Exists         *bool        `json:"exists,omitempty"`   // only exists request
}

type RoomReq struct {
	Request         string      `json:"request"`    // "destroy|exists|listparticipants"
	Room            float64     `json:"room"`
}

type RoomKick struct {
	Request         string      `json:"request"`    // "create"
	Room            float64     `json:"room"`       // <unique numeric ID, optional, chosen by plugin if missing>
	Id              float64     `json:"id"`         // unique numeric ID of the participant to kick
}

type JanusRoomParticipants struct {
	VideoRoom       string      `json:"videoroom"`  // "participants"
	Room            float64     `json:"room"`
	Participants    []JanusParticipant  `json:"participants"`
}
type JanusParticipant struct {
	Id          float64     `json:"id"`
	Display     string      `json:"display"`
	Publisher   string      `json:"publisher"`
	Talking     bool        `json:"talking"`
}

type RoomDestroy struct {
	Request         string      `json:"request"`    // "destroy"
	Room            float64     `json:"room"`       // <unique numeric ID of the room to destroy>
}