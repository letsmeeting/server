package lilpop

import (
	"encoding/json"
	"errors"
	ws "github.com/jinuopti/lilpop-server/communication/http/websocket"
	"github.com/jinuopti/lilpop-server/configure"
	"github.com/jinuopti/lilpop-server/extension/janus"
	utility "github.com/jinuopti/lilpop-server/library"
	. "github.com/jinuopti/lilpop-server/log"
	"strings"
)

var (
	MatchMap map[string]*MatchInfo
)

type MatchInfo struct {
	userId          string

	appClient 		*ws.Client

	gateway         *janus.Gateway
	session			*janus.Session

	videoHandle 	*janus.Handle
	textHandle		*janus.Handle

	videoPeerHandle *janus.Handle
	textPeerHandle  *janus.Handle

	room            *Room
	category 		[]string

	enterNo         int         // 방 입장 순서
	roomNo          float64     // 방 번호
	privateId       float64
	publisherId     float64

	publishers      []janus.PublisherInfo       // 대화방에 있는 user 정보 (janus)
	peerUserId      []string                    // 대화방에 있는 user id slice

	OfferReady      bool    // 상대에게 보내는 sdp 완료
	AnswerReady     bool    // 내 앱에 보이는 sdp 완료
	Completed       bool    // 모든 절차 완료

	prevTransaction string
	startData       *StartBody      // category, region, tag

	keepAlive       bool            // janus keepalive signaling
	exit            chan bool       // 종료
}

func InitMatch() {
	if MatchMap == nil {
		MatchMap = make(map[string]*MatchInfo)
	}
}

func Detach(userId string) {
	j := MatchMap[userId]
	if j == nil {
		return
	}
	j.exit <- true

	_ = j.room.LeavingRoom(j)
	_ = j.room.DestroyRoom(j)

	if j.videoHandle != nil {
		_, _ = j.videoHandle.Detach()
	}
	if j.session != nil {
		_, _ = j.session.Destroy()
	}

	delete(MatchMap, userId)
}

func NewJanusInfo(userId string) *MatchInfo {
	if MatchMap[userId] != nil {
		return MatchMap[userId]
	}
	info := &MatchInfo{
		userId: userId,
	}
	info.exit = make(chan bool, 1)
	MatchMap[userId] = info

	return info
}

func CreateSession(gw *janus.Gateway) (*janus.Session, error) {
	session, err := gw.Create()
	if err != nil || session == nil {
		Logd("error, %s", err)
		return nil, err
	}
	return session, nil
}

func CreateHandleVideoroom(session *janus.Session) (*janus.Handle, error) {
	handle, err := session.Attach("janus.plugin.videoroom")
	if err != nil || handle == nil {
		Logd("error, %s", err)
		return nil, err
	}
	return handle, nil
}

func PrepareJoinRoom(c *ws.Client, data map[string]interface{}) (*MatchInfo, error) {
	config := configure.GetConfig()

	InitMatch()
	// connect
	gw, _ := janus.Connect(config.Lilpop.JanusAddr, config.Lilpop.JanusWebsocketPort)

	d := data["data"].(map[string]interface{})
	category := d["category"].(string)
	region := d["region"].(string)
	tag := d["tag"].(string)
	Logd("Category:[%s], Region:[%s], Tag:[%s]", category, region, tag)

	// create channel
	Logd("Janus Create")
	session, err := CreateSession(gw)
	if err != nil {
		return nil, err
	}

	// attach channel
	Logd("Janus Attach")
	handle, err := CreateHandleVideoroom(session)
	if err != nil {
		Logd("error, %s", err)
		return nil, err
	}

	j := NewJanusInfo(c.UserId)
	j.gateway = gw
	j.session = session
	j.videoHandle = handle
	j.appClient = c
	j.publishers = []janus.PublisherInfo{}
	j.enterNo = len(j.publishers) + 1
	j.prevTransaction = data["transaction"].(string)
	startBody := data["data"].(map[string]interface{})
	j.startData = &StartBody{
		Category: startBody["category"].(string),
		Region: startBody["region"].(string),
		Tag: startBody["tag"].(string),
	}
	j.category = utility.DeleteEmpty(strings.Split(category, ","))
	if len(j.category) == 0 {
		j.category = append(j.category, TestRoom)
	}

	Logd("UserId:[%s] handle_id=[%d], publishers:[%d], enterNo[%d], category[%d][%v]",
		c.UserId, handle.ID, len(j.publishers), j.enterNo, len(j.category), j.category)

	// TODO: Matching algorithm - get room id
	_, err = GetRoom(j)
	if err != nil {
		j.roomNo = 1234
	}

	go j.Events()

	// join room
	err = j.JoinRoom("publisher")
	if err != nil {
		return j, err
	}

	return j, nil
}

func (j *MatchInfo) JoinRoom(ptype string) error {
	joinMsg := janus.JoinMsg{
		Request: "join",
		Room: j.roomNo,
		Ptype: "publisher",
		Display: j.userId,
	}

	var handle *janus.Handle
	if ptype == "subscriber" {
		handle = j.videoPeerHandle
	} else {
		handle = j.videoHandle
	}

	Logd("Join Room No:%f ptype:%s, handle:%d", joinMsg.Room, joinMsg.Ptype, handle.ID)

	eMsg, err := handle.Message(joinMsg, nil)
	if err != nil {
		Logd("error, %s", err)
		return err
	}
	_ = j.EventProcess(eMsg)

	return nil
}

func (j *MatchInfo) NextMatch(c *ws.Client, jsonBody map[string]interface{}) error {
	top := make(map[string]interface{})
	body := make(map[string]interface{})
	body["category"] = j.category[0]
	body["region"] = ""
	body["tag"] = ""
	top["data"] = body
	top["transaction"] = jsonBody["transaction"].(string)

	_ = j.room.LeavingRoom(j)
	_ = j.room.DestroyRoom(j)

	if j.videoHandle != nil {
		_, _ = j.videoHandle.Detach()
	}
	if j.session != nil {
		_, _ = j.session.Destroy()
	}

	newInfo, err := PrepareJoinRoom(c, top)
	if err != nil {
		return err
	}
	if len(newInfo.publishers) > 0 {
		return nil
	}
	request := RequestReq{Lilpop: "request", Type: "offer"}
	sMessage, err := json.Marshal(request) // request
	c.Send <- sMessage

	return nil
}

func Offer(userId string, audio bool, video bool, sdpType string, sdp string) (map[string]interface{}, error) {
	if MatchMap[userId] == nil {
		Logd("userId %s map is nil", userId)
		return nil, errors.New("not found user " + userId)
	}
	info := MatchMap[userId]
	vHandle := info.videoHandle

	//tHandle := info.textHandle
	if vHandle == nil {
		Logd("Video handle is nil")
		return nil, errors.New("video handle is nil")
	}

	body := janus.OfferBody{
		Request: "configure",
		Audio: audio,
		Video: video,
	}
	jsep := janus.OfferJsep{
		Type: sdpType,
		Sdp: sdp,
	}

	eventMsg, err := vHandle.Message(body, jsep)
	if err != nil {
		Logd("error, %s", err)
		return nil, err
	}

	answer, err := MakeAnswer(eventMsg)
	if err != nil {
		Logd("error, %s", err)
		return nil, err
	}
	return answer, nil
}

func MakeAnswer(answer *janus.EventMsg) (map[string]interface{}, error) {
	if answer.Plugindata.Plugin == "janus.plugin.videoroom" {
		return MakeAnswerVideo(answer)
	} else if answer.Plugindata.Plugin == "janus.plugin.textroom" {
		return MakeAnswerText(answer)
	}
	return nil, errors.New("unknown plugin " + answer.Plugindata.Plugin)
}

func MakeAnswerVideo(answer *janus.EventMsg) (map[string]interface{}, error) {
	pluginData := janus.PluginVideoData{
		VideoRoom: answer.Plugindata.Data["videoroom"].(string),
		Room: answer.Plugindata.Data["room"].(float64),
		Configured: answer.Plugindata.Data["configured"].(string),
	}
	if answer.Plugindata.Data["audio_codec"] != nil {
		pluginData.AudioCodec = answer.Plugindata.Data["audio_codec"].(string)
	}
	if answer.Plugindata.Data["video_codec"] != nil {
		pluginData.VideoCodec = answer.Plugindata.Data["video_codec"].(string)
	}
	jsep := janus.Jsep{
		Type: answer.Jsep["type"].(string),
		Sdp: answer.Jsep["sdp"].(string),
	}

	rsp := make(map[string]interface{})
	if len(pluginData.AudioCodec) > 0 {
		rsp["audio_codec"] = pluginData.AudioCodec
	}
	if len(pluginData.VideoCodec) > 0 {
		rsp["video_codec"] = pluginData.VideoCodec
	}
	rsp["sdp"] = jsep.Sdp

	return rsp, nil
}

func MakeAnswerText(answer *janus.EventMsg) (map[string]interface{}, error) {
	rsp := make(map[string]interface{})

	return rsp, nil
}

func Trickle(userId string, candidate interface{}) error {
	info := MatchMap[userId]
	if info == nil {
		return errors.New("not found user map")
	}

	sMsg := make(map[string]interface{})
	candidateMap := candidate.(map[string]interface{})
	if candidateMap["candidate"].(string) == "completed" {
		Logd("Trickle complete")
		sMsg["completed"] = true
	} else {
		sMsg = candidateMap
	}

	_, err := info.videoHandle.Trickle(sMsg)
	if err != nil {
		return errors.New("trickle error")
	}

	return nil
}

func (j *MatchInfo) MatchReady(isOffer bool) error {
	if j.Completed {
		Logd("Already completed")
		return nil
	}

	isRequest := false
	var packet interface{}
	var ready SimpleReq
	if isOffer {
		if j.AnswerReady {
			ready.Lilpop = CmdComplete
			j.Completed = true
		} else if j.enterNo == 1 {
			ready.Lilpop = CmdReady
		} else {
			return nil
		}
	} else {
		if j.OfferReady {
			ready.Lilpop = CmdComplete
			j.Completed = true
		} else {
			rsp := RequestReq{
				Lilpop: CmdRequest,
				Type: "offer",
			}
			packet = rsp
			isRequest = true
		}
	}

	if isRequest == false {
		packet = ready
	}

	sMessage, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	j.appClient.Send <- sMessage

	if ready.Lilpop == CmdComplete || ready.Lilpop == CmdReady {
		if !j.keepAlive {
			go j.KeepAlive()
		}
	}

	return nil
}

func (j *MatchInfo) Hangup(e *janus.HangupMsg) error {
	var event EventServer

	if e.Handle == j.videoHandle.ID {
		Logd("Hangup me? id:%d", j.videoHandle.ID)
	} else if e.Handle == j.videoPeerHandle.ID {
		if e.Reason == "DTLS alert" {
			Logd("peer hangup, but DTLS alert")
			return nil
		}
		event.Lilpop = CmdEvent
		event.UserId = j.publishers[0].Display
		event.Type = "hangup"
		event.Value = "1"
		sMessage, err := json.Marshal(event)
		if err != nil {
			return err
		}
		j.appClient.Send <- sMessage
	} else {
		Logd("unknown handle id e:%d, me:%d, peer:%d", e.Handle, j.videoHandle.ID, j.videoPeerHandle.ID)
	}

	return nil
}