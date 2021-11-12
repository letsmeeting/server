package lilpop

import (
	"encoding/json"
	"errors"
	ws "github.com/jinuopti/lilpop-server/communication/http/websocket"
	"github.com/jinuopti/lilpop-server/database/gorm/userdb"
	"github.com/jinuopti/lilpop-server/extension/janus"
	utility "github.com/jinuopti/lilpop-server/library"
	. "github.com/jinuopti/lilpop-server/log"
	token "github.com/jinuopti/lilpop-server/restapi/v1/user"
	"github.com/labstack/echo/v4"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

func Recover(c *ws.Client) {
	if v := recover(); v != nil {
		function, file, line, _ := runtime.Caller(3)
		Loge("Recovered, Caller %s:%d (%s), Message: %s",
			filepath.Base(file), line, runtime.FuncForPC(function).Name(), v)
		Loge("stacktrace from panic: \n" + string(debug.Stack()))
		c.Disconnect = true
	}
}

func WsAppHandler(c echo.Context) error {
	err := ws.WebSocketHandler(c, WsRead, CloseCallback)
	if err != nil {
		return err
	}

	return nil
}

func CloseCallback(c *ws.Client) {
	Logd("UserId: %s, Close janus gateway websocket", c.UserId)
	Detach(c.UserId)
}

func WsRead(c *ws.Client, message []byte) {
	defer Recover(c)

	prettyJson, err := utility.GetPrettyJsonStr(message)
	if err == nil {
		Logd("Lilpop Websocket Read: \n%s", prettyJson)
	}

	// message type 에 따른 server push
	jsonBody := make(map[string]interface{})
	err = json.Unmarshal(message, &jsonBody)
	if err != nil {
		Logd("error: %s", err)
		return
	}

	var errMessage string
	var sMessage []byte

	req := &GeneralReq{
		Lilpop: jsonBody["lilpop"].(string),
		Transaction: jsonBody["transaction"].(string),
	}

	// Login 외 API는 Login 하지 않으면 에러
	if req.Lilpop != CmdLogin {
		if !c.Login {
			Logd("Not logged in, user %s", c.UserId)
			errMessage = "Not logged in"
			SendErrorMessage(c, req.Transaction, errMessage)
			return
		}
	}

	switch req.Lilpop {
	case CmdLogin:
		userId, err := ProcCmdLogin(jsonBody)
		if err != nil {
			errMessage = err.Error()
			SendErrorMessage(c, req.Transaction, errMessage)
			return
		} else {
			rsp := LoginRsp{
				Lilpop: "success",
				Transaction: req.Transaction,
			}
			sMessage, err = json.Marshal(rsp)
			c.Login = true
			c.UserId = userId
		}

	case CmdStart:
		Logd("Match Start")
		j, err := PrepareJoinRoom(c, jsonBody)
		if err != nil {
			errMessage = err.Error()
			SendErrorMessage(c, req.Transaction, errMessage)
			return
		}
		if len(j.publishers) > 0 {
			return
		}
		request := RequestReq{Lilpop: "request", Type: "offer"}
		sMessage, err = json.Marshal(request) // request

	case CmdNext:
		Logd("Next Match")
		rsp := GeneralRsp{
			Lilpop: "success",
			Transaction: req.Transaction,
			Message: "success request next matching",
		}
		j := MatchMap[c.UserId]
		failNext := false
		if j == nil || j.room == nil {
			failNext = true
		} else {
			sMessage, err = json.Marshal(rsp)
			c.Send <- sMessage
			err = j.NextMatch(c, jsonBody)
			if err != nil {
				failNext = true
			} else {
				return
			}
		}
		if failNext {
			rsp.Lilpop = "fail"
			rsp.Message = "fail request next matching"
		}
		sMessage, err = json.Marshal(rsp)

	case CmdExit:
		Logd("Exit, close websocket")
		Detach(c.UserId)
		c.Exit <- true
		c.World.ChanLeave <- c
		_ = c.Conn.Close()
		return

	case CmdOffer:
		Logd("Offer Request")
		r, err := ProcCmdSdpApp(c.UserId, jsonBody)
		rsp := r.(*SdpServerAnswer)
		if err != nil {
			errMessage = err.Error()
			SendErrorMessage(c, req.Transaction, errMessage)
			return
		}
		sMessage, err = json.Marshal(rsp)

	case CmdAnswer:
		Logd("Answer Request")
		r, err := ProcCmdSdpApp(c.UserId, jsonBody)
		if r == nil {
			return
		}
		rsp := r.(*RequestReq)
		if err != nil {
			errMessage = err.Error()
			SendErrorMessage(c, req.Transaction, errMessage)
			return
		}
		sMessage, err = json.Marshal(rsp)

	case CmdTrickle:
		Logd("Trickle")
		err = Trickle(c.UserId, jsonBody["candidate"])
		if err != nil {
			errMessage = err.Error()
			SendErrorMessage(c, req.Transaction, errMessage)
			return
		}
		rsp := GeneralReq{
			Lilpop: CmdAck,
			Transaction: req.Transaction,
		}
		sMessage, err = json.Marshal(rsp)

	case CmdEvent:
		Logd("Event")
		ProcCmdEvent(jsonBody["type"].(string), jsonBody["value"].(string))

	default:
		Logd("error, unknown command [%s]", req.Lilpop)
		errMessage = "Unknown Lilpop Command " + req.Lilpop
		SendErrorMessage(c, req.Transaction, errMessage)
		return
	}

	c.Send <- sMessage
}

func SendErrorMessage(c *ws.Client, transaction string, errMessage string) {
	rspFail := GeneralRsp{
		Lilpop: CmdFail,
		Transaction: transaction,
		Message: errMessage,
	}
	sMessage, err := json.Marshal(rspFail)
	if err != nil {
		Logd("err, %s", err)
		return
	}
	c.Send <- sMessage
	c.Disconnect = true
}

func ProcCmdLogin(json map[string]interface{}) (string, error) {
	var err error
	var loginReq LoginReq

	loginReq.Lilpop = json["lilpop"].(string)
	loginReq.Transaction = json["transaction"].(string)
	body := json["data"].(map[string]interface{})
	loginReq.Data.UserId = body["user_id"].(string)
	loginReq.Data.AccessToken = body["access_token"].(string)
	user := userdb.FindByIdUser(loginReq.Data.UserId)
	if user == nil {
		Logd("User does not exists [%s]", loginReq.Data.UserId)
		return "", errors.New("user does not exist")
	} else {
		// verify access_token
		_, err = token.VerifyToken(loginReq.Data.UserId, "access_token", loginReq.Data.AccessToken)
		if err != nil {
			return "", err
		}
	}
	return loginReq.Data.UserId, nil
}

func ProcCmdSdpApp(userId string, json map[string]interface{}) (interface{}, error) {
	var sdpApp SdpApp

	sdpApp.Lilpop = json["lilpop"].(string)
	sdpApp.Transaction = json["transaction"].(string)
	body := json["data"].(map[string]interface{})
	if body["audio"] != nil {
		sdpApp.Data.Audio = body["audio"].(bool)
	} else {
		sdpApp.Data.Audio = true
	}
	if body["video"] != nil {
		sdpApp.Data.Video = body["video"].(bool)
	} else {
		sdpApp.Data.Video = true
	}

	sdp := body["sdp"].(map[string]interface{})
	sdpApp.Data.Sdp.Type = strings.ToLower(sdp["type"].(string))
	sdpApp.Data.Sdp.Description = sdp["description"].(string)

	switch sdpApp.Data.Sdp.Type {
	case "offer":
		answer, err := Offer(userId, sdpApp.Data.Audio, sdpApp.Data.Video,
			sdpApp.Data.Sdp.Type, sdpApp.Data.Sdp.Description)
		if err != nil {
			return nil, err
		} else if answer == nil {
			return nil, nil
		}
		rsp := &SdpServerAnswer{
			Lilpop: CmdAnswer,
			Transaction: sdpApp.Transaction,
		}
		rsp.Data.AudioCodec = answer["audio_codec"].(string)
		rsp.Data.VideoCodec = answer["video_codec"].(string)
		rsp.Data.Sdp = answer["sdp"].(string)
		return rsp, nil

	case "answer":
		info := MatchMap[userId]
		vHandle := info.videoPeerHandle
		body := janus.AnswerBody{
			Request: "start",
			Room: info.roomNo,
		}
		jsep := janus.OfferJsep{
			Type: "answer",
			Sdp: sdpApp.Data.Sdp.Description,
		}
		eMsg, err := vHandle.Message(body, jsep)
		if err != nil {
			Logd("error, %s", err)
			return nil, err
		}
		pluginData := janus.PluginVideoDataStarted{
			VideoRoom: eMsg.Plugindata.Data["videoroom"].(string),
			Room: eMsg.Plugindata.Data["room"].(float64),
			Started: eMsg.Plugindata.Data["started"].(string),
		}
		Logd("Started %s", pluginData.Started)

		if pluginData.Started == "ok" {
			go info.EventsYou()
			return nil, nil
		}
		return nil, errors.New("started command is fail")

	default:
		return nil, errors.New("Invalid sdp type " + sdpApp.Data.Sdp.Type)
	}
}

func ProcCmdEvent(key string, value string) {

}