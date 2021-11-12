package lilpop

import (
    "encoding/json"
    "github.com/jinuopti/lilpop-server/extension/janus"
    . "github.com/jinuopti/lilpop-server/log"
    "time"
)

func (j *MatchInfo) KeepAlive() {
    ticker := time.NewTicker(time.Second * time.Duration(30))
    j.keepAlive = true

    for {
        select {
        case <-j.exit:
            j.keepAlive = false
            Logd("Exit KeepAlive goroutine (%s)", j.userId)
            return
        case <-ticker.C:
            _, _ = j.session.KeepAlive()
        }
    }
}

func (j *MatchInfo) Events() {
    for {
        c, ok := <-j.videoHandle.Events
        if !ok {
            Logd("It's not ok")
            return
        }
        switch message := c.(type) {
        case *janus.EventMsg:
            Logd("event: %s", message)
            _ = j.EventProcess(c)
        case *janus.WebRTCUpMsg:
            Logd("webrtcup: %s", message)
            j.OfferReady = true
            _ = j.MatchReady(true)
        case *janus.HangupMsg:
            Logd("hangup: %s", message)
            _ = j.Hangup(c.(*janus.HangupMsg))
        case *janus.DetachedMsg:
            Logd("detached: %s", message)
        case *janus.MediaMsg:
            Logd("media: %s", message)
        case *janus.SlowLinkMsg:
            Logd("slow link: %s", message)
        default:
            Logd("Unknown message: %v", c)
        }
    }
}

func (j *MatchInfo) EventsYou() {
    for {
        c, ok := <-j.videoPeerHandle.Events
        if !ok {
            Logd("It's not ok")
            return
        }
        switch message := c.(type) {
        case *janus.EventMsg:
            Logd("event: %s", message)
            _ = j.EventProcess(c)
        case *janus.WebRTCUpMsg:
            Logd("webrtcup: %s", message)
            j.AnswerReady = true
            _ = j.MatchReady(false)
        case *janus.HangupMsg:
            Logd("hangup: %s", message)
            _ = j.Hangup(c.(*janus.HangupMsg))
        case *janus.DetachedMsg:
            Logd("detached: %s", message)
        case *janus.MediaMsg:
            Logd("media: %s", message)
        case *janus.SlowLinkMsg:
            Logd("slow link: %s", message)
        default:
            Logd("Unknown message: %v", c)
        }
    }
}

func (j *MatchInfo) EventProcess(event interface{}) error {
    defer Recover(j.appClient)

    e := event.(*janus.EventMsg)
    pluginData := e.Plugindata

    if pluginData.Plugin == "janus.plugin.videoroom" {
        videoroom := e.Plugindata.Data["videoroom"].(string)
        Logd("videoroom: [%s]", videoroom)

        switch videoroom {
        case "slow_link":
            Logd("slow_link")
            return nil
        }

        room := e.Plugindata.Data["room"].(float64)
        Logd("room: %v", room)

        if e.Plugindata.Data["publishers"] != nil {
            Logd("PluginData: publishers")
            _ = j.EventPublishers(e)
        } else if e.Plugindata.Data["attached"] != nil {
            j.EventAttached(e)
        } else if e.Plugindata.Data["started"] != nil {
            Logd("PluginData: started")
            j.EventStarted(e)
        } else if e.Plugindata.Data["unpublished"] != nil {
            Logd("PluginData: publishers")
        } else if e.Plugindata.Data["leaving"] != nil {
            Logd("PluginData: leaving")
        } else if e.Plugindata.Data["configured"] != nil {
            Logd("PluginData: configured")
        } else if e.Plugindata.Data["error"] != nil {
            Logd("PluginData error, [%f]%s",
                e.Plugindata.Data["error_code"].(float64), e.Plugindata.Data["error"].(string))
        }
    }

    return nil
}

func (j *MatchInfo) EventPublishers(e *janus.EventMsg) error {
    videoroom := e.Plugindata.Data["videoroom"].(string)
    Logd("videoroom: [%s]", videoroom)
    room := e.Plugindata.Data["room"].(float64)
    Logd("room: %v", room)

    if j.privateId <= 0 {
        j.privateId = e.Plugindata.Data["private_id"].(float64)
    }

    publishers := e.Plugindata.Data["publishers"].([]interface{})
    for i, pubs := range publishers {
        var p janus.PublisherInfo
        pub := pubs.(map[string]interface{})
        p.Id = pub["id"].(float64)
        p.Display = pub["display"].(string)
        p.AudioCodec = pub["audio_codec"].(string)
        p.VideoCodec = pub["video_codec"].(string)
        p.Talking = pub["talking"].(bool)

        j.publishers = append(j.publishers, p)

        Logd("[%d] Room:[%v] Id:[%v], Display:[%s], AudioCodec:[%s], VideoCodec:[%s], Talking:[%v]",
            i, room, p.Id, p.Display, p.AudioCodec, p.VideoCodec, p.Talking)
    }

    rsp := JoinedRsp{
        Lilpop:      CmdJoined,
        Transaction: j.prevTransaction,
        Publishers: len(j.publishers),
    }
    sMessage, err := json.Marshal(rsp)
    j.appClient.Send <- sMessage // joined

    if len(publishers) == 0 {
        return nil
    }

    Logd("Janus Attach")
    handle, err := j.session.Attach("janus.plugin.videoroom")
    if err != nil || handle == nil {
        Logd("error, %s", err)
        return err
    }
    j.videoPeerHandle = handle

    // join room
    joinMsg := janus.JoinSubscribe{
        Request: "join",
        Room:    j.roomNo,
        Ptype:   "subscriber",
        PrivateId: j.privateId,
        Feed: j.publishers[0].Id,
    }
    Logd("Janus Message (join)")
    offer, err := j.videoPeerHandle.Message(joinMsg, nil)
    if err != nil {
        Logd("error, %s", err)
        return err
    }
    j.EventAttached(offer)

    return nil
}

func (j *MatchInfo) EventAttached(e *janus.EventMsg) {
    Logd("plugin:[%s], cmd:[%s]", e.Plugindata.Plugin, e.Plugindata.Data["videoroom"].(string))

    pluginData := janus.VideoPluginDataEventAttached{
        Display: e.Plugindata.Data["display"].(string),
        Id: e.Plugindata.Data["id"].(float64),
        Room: e.Plugindata.Data["room"].(float64),
        Videoroom: e.Plugindata.Data["videoroom"].(string),
    }
    jsep := janus.Jsep{
        Type: e.Jsep["type"].(string),
        Sdp: e.Jsep["sdp"].(string),
    }

    var packet SdpServerOffer
    packet.Lilpop = CmdOffer
    packet.UserId = pluginData.Display
    packet.Data.VideoCodec = j.publishers[0].VideoCodec
    packet.Data.AudioCodec = j.publishers[0].AudioCodec
    packet.Data.Sdp = jsep.Sdp

    sMessage, err := json.Marshal(packet)
    if err != nil {
        Loge("%v", err)
        return
    }

    j.appClient.Send <- sMessage
}

func (j *MatchInfo) EventStarted(e *janus.EventMsg) {

}