package lilpop

import (
    "errors"
    "github.com/go-redis/redis/v8"
    "github.com/jinuopti/lilpop-server/database/redisdb"
    "github.com/jinuopti/lilpop-server/extension/janus"
    . "github.com/jinuopti/lilpop-server/log"
    "strconv"
    "strings"
    "sync"
)

const (
    RedisPrefixRoom = "room"
)

const (
    Music = "음악"
    Acting = "연기"
    Magic = "마술"
    Art = "미술"
    Gag = "개그"
    Dance = "댄스"
    Guitar = "기타"
    Keyboard = "건반"
    Vocal = "보컬"
    Bass = "베이스"
    Drum = "드럼"
    Windwood = "윈드우드"
    Brass = "브라스"
    String = "스트링"
    Composition = "작편곡"
    KTM = "국악"

    TestRoom = "테스트룸"
)

const (
    DefaultPublishers = 2
    DefaultVideoCodec = "vp8"
    DefaultAudioCodec = "opus"
    DefaultBitRate = 2048000
    DefaultRecord = false
    DefaultTitle = "즐거운 만남"
)

const (
    RoomNoOffset = 10000

    MusicRoomNo     = RoomNoOffset * 1
    ActingRoomNo    = RoomNoOffset * 2
    MagicRoomNo     = RoomNoOffset * 3
    ArtRoomNo       = RoomNoOffset * 4
    GagRoomNo       = RoomNoOffset * 5
    DanceRoomNo     = RoomNoOffset * 6
    GuitarRoomNo    = RoomNoOffset * 7
    KeyboardRoomNo  = RoomNoOffset * 8
    VocalRoomNo     = RoomNoOffset * 9
    BassRoomNo      = RoomNoOffset * 10
    DrumRoomNo      = RoomNoOffset * 11
    WindwoodRoomNo  = RoomNoOffset * 12
    BrassRoomNo     = RoomNoOffset * 13
    StringRoomNo    = RoomNoOffset * 14
    CompositionRoomNo = RoomNoOffset * 15
    KtmRoomNo       = RoomNoOffset * 16
)

type Room struct {
    Id          uint64      // janus room id

    Type        string      // "음악"
    Title       string      // "음악을 사랑하는 사람들"

    MaxUsers    float64     // 방 최대 입장 제한 (default: 8)
    CurrUsers   float64     // 현재 방 설정 인원
    JoinUsers   float64     // 현재 방 입장 인원

    UserId      []string    // 방에 입장한 user id list
}

var (
    // RoomMap = make(map[uint64]*Room)

    mutex = &sync.Mutex{}

    MusicNo uint64 = MusicRoomNo
    ActingNo uint64 = ActingRoomNo
    MagicNo uint64 = MagicRoomNo
    ArtNo uint64 = ArtRoomNo
    GagNo uint64 = GagRoomNo
    DanceNo uint64 = DanceRoomNo
    GuitarNo uint64 = GuitarRoomNo
    KeyboardNo uint64 = KeyboardRoomNo
    VocalNo uint64 = VocalRoomNo
    BassNo uint64 = BassRoomNo
    DrumNo uint64 = DrumRoomNo
    WindwoodNo uint64 = WindwoodRoomNo
    BrassNo uint64 = BrassRoomNo
    StringNo uint64 = StringRoomNo
    CompositionNo uint64 = CompositionRoomNo
    KtmNo uint64 = KtmRoomNo
)

func GetRoomId(category string, prevRoomNo uint64) (uint64, bool, *Room) {
    var roomId uint64

    value, min, max := getMinMaxValue(category)
    Logd("value:%v, min:%v, max:%v", *value, min, max)
    mutex.Lock()
    defer mutex.Unlock()

    start := *value
    found := false

    r := &Room{}
    for {
        key := RedisPrefixRoom + ":" + strconv.Itoa(int(*value))
        err := redisdb.GetStruct(key, r)
        if err == redis.Nil {   // empty room
            found = true
            break
        } else if err != nil {
            Loge("error, %s", err)
        } else {
            Logd("Room id:%d is not null, join:%v, max:%v", *value, r.JoinUsers, r.MaxUsers)
            // 기존 방에 자리가 있어 입장
            if r.JoinUsers < r.MaxUsers && prevRoomNo != *value {
                roomId = *value
                Logd("found previous room %d", roomId)
                return roomId, false, r
            }
            *value = *value + 1
            if *value >= max {
                *value = min
            }
        }
        if start == *value {
            break
        }
    }
    if found {
        roomId = *value
        Logd("found new room %d", roomId)
    }

    // 신규 방 생성
    return roomId, true, nil
}

func (r *Room) LeavingRoom(j *MatchInfo) error {
    leave := &janus.RoomSimpleReq{
        Request: "leave",
    }
    Logd("Janus Message (leave room)")
    e, err := j.videoHandle.Message(leave, nil)
    if err != nil {
        Logd("error, %s", err)
        return err
    }
    pluginData := e.Plugindata
    if pluginData.Plugin != "janus.plugin.videoroom" {
        Loge("[%s] not videoroom", pluginData.Plugin)
        return errors.New("invalid plugin")
    }
    videoroom := e.Plugindata.Data["videoroom"].(string)
    if videoroom != "event" {
        return errors.New("invalid videoroom response " + videoroom)
    }
    if e.Plugindata.Data["error_code"] != nil {
        errorMsg := e.Plugindata.Data["error"].(string)
        Loge("error_code=%d, msg: %s", e.Plugindata.Data["error_code"].(float64), errorMsg)
    }
    if e.Plugindata.Data["leaving"] != nil {
        ok := e.Plugindata.Data["leaving"].(string)
        Logd("success leaving room %s, JoinUsers=%d", ok, uint64(r.JoinUsers))
    }
    r.JoinUsers = r.JoinUsers - 1

    return nil
}

func (r *Room) DestroyRoom(j *MatchInfo) error {
    if r.Type == TestRoom {
        return nil
    }
    if r.JoinUsers > 0 {
        return nil
    }
    destroy := &janus.RoomDestroy{
        Request: "destroy",
        Room: float64(r.Id),
    }
    Logd("Janus Message (destroy room)")
    e, err := j.videoHandle.MessageCreate(destroy)
    if err != nil {
        Logd("error, %s", err)
        return err
    }
    pluginData := e.PluginData
    if pluginData.Plugin != "janus.plugin.videoroom" {
        Loge("[%s] not videoroom", pluginData.Plugin)
        return errors.New("invalid plugin")
    }
    videoroom := e.PluginData.Data["videoroom"].(string)
    if videoroom != "destroyed" {
        return errors.New("invalid videoroom response " + videoroom)
    }
    room := int64(e.PluginData.Data["room"].(float64))

    // delete(RoomMap, r.Id)
    key := RedisPrefixRoom + ":" + strconv.Itoa(int(r.Id))
    c := redisdb.Del(key)

    Logd("success destroy room %d, redis del %d", room, c)

    return nil
}

func GetRoom(j *MatchInfo) (*Room, error) {
    // TEST
    if j.category[0] == TestRoom || j.category[0] == Composition {
        testRoom := &Room{
            Id: 1234,
            Type: TestRoom,
            Title: "테스트 룸 입니다.",
            MaxUsers: 8,
            CurrUsers: 8,
            JoinUsers: 0,
        }
        testRoom.UserId = append(testRoom.UserId, j.userId)
        testRoom.JoinUsers = testRoom.JoinUsers + 1
        j.room = testRoom
        j.roomNo = float64(testRoom.Id)
        return nil, errors.New("set room 1234 for test")
    }

    // var r *Room
    id, isCreate, r := GetRoomId(j.category[0], uint64(j.roomNo))
    if isCreate {
        r = &Room {
            Id: id,
            Type: j.category[0],
            Title: DefaultTitle,
            MaxUsers: DefaultPublishers,
            CurrUsers: DefaultPublishers,
            JoinUsers: 0,
        }
        r.UserId = make([]string, DefaultPublishers)
        // RoomMap[id] = r
    }

    if isCreate {
        err := r.CreateRoom(j)
        if err != nil {
            if !strings.Contains(err.Error(), "already exists") {
                // delete(RoomMap, id)
                // Logd("error, %s", err)
                // return nil, err
                Logd("room %d is already exists, join room", r.Id)
            }
        }
    }
    r.UserId = append(r.UserId, j.userId)
    r.JoinUsers = r.JoinUsers + 1
    j.room = r
    j.roomNo = float64(r.Id)

    if isCreate {
        key := RedisPrefixRoom + ":" + strconv.Itoa(int(id))
        _, err := redisdb.SetNX(key, r, 0)
        if err != nil {
            Loge("redis set error, %s", err)
            return nil, err
        }
    }

    Logd("Room [%v], Max:%v, Join:%v, Curr:%v, UserId[%v]", r.Id, r.MaxUsers, r.JoinUsers, r.CurrUsers, r.UserId)

    return r, nil
}

func (r *Room) CreateRoom(j *MatchInfo) error {
    createRoom := &janus.RoomCreateReq{
        Request: "create",
        Room: float64(r.Id),
        Publishers: r.MaxUsers,
        Description: r.Title,
        VideoCodec: DefaultVideoCodec,
        BitRate: DefaultBitRate,
        Record: DefaultRecord,
    }
    Logd("Janus Message (create room)")
    e, err := j.videoHandle.MessageCreate(createRoom)
    if err != nil {
        Logd("error, %s", err)
        return err
    }

    pluginData := e.PluginData
    if pluginData.Plugin != "janus.plugin.videoroom" {
        Loge("[%s] not videoroom", pluginData.Plugin)
        return errors.New("invalid plugin")
    }
    videoroom := e.PluginData.Data["videoroom"].(string)
    if videoroom == "event" {
        code := e.PluginData.Data["error_code"].(float64)
        errMsg := e.PluginData.Data["error"].(string)
        Loge("event error_code[%d], error: %s", code, errMsg)
        return errors.New(errMsg)
    }
    if videoroom != "created" {
        return errors.New("invalid videoroom response " + videoroom)
    }

    room := int64(e.PluginData.Data["room"].(float64))
    permanent := e.PluginData.Data["permanent"].(bool)
    Logd("success create room %d, permanent=%v", room, permanent)

    return nil
}

func getMinMaxValue(category string) (value *uint64, min uint64, max uint64) {
    switch category {
    case Music:
        value = &MusicNo
        min = MusicRoomNo
        max = MusicRoomNo + RoomNoOffset
    case Acting:
        value = &ActingNo
        min = ActingRoomNo
        max = ActingRoomNo + RoomNoOffset
    case Magic:
        value = &MagicNo
        min = MagicRoomNo
        max = MagicRoomNo + RoomNoOffset
    case Art:
        value = &ArtNo
        min = ArtRoomNo
        max = ArtRoomNo + RoomNoOffset
    case Gag:
        value = &GagNo
        min = GagRoomNo
        max = GagRoomNo + RoomNoOffset
    case Dance:
        value = &DanceNo
        min = DanceRoomNo
        max = DanceRoomNo + RoomNoOffset
    case Guitar:
        value = &GuitarNo
        min = GuitarRoomNo
        max = GuitarRoomNo + RoomNoOffset
    case Keyboard:
        value = &KeyboardNo
        min = KeyboardRoomNo
        max = KeyboardRoomNo + RoomNoOffset
    case Vocal:
        value = &VocalNo
        min = VocalRoomNo
        max = VocalRoomNo + RoomNoOffset
    case Bass:
        value = &BassNo
        min = BassRoomNo
        max = BassRoomNo + RoomNoOffset
    case Drum:
        value = &DrumNo
        min = DrumRoomNo
        max = DrumRoomNo + RoomNoOffset
    case Windwood:
        value = &WindwoodNo
        min = WindwoodRoomNo
        max = WindwoodRoomNo + RoomNoOffset
    case Brass:
        value = &BrassNo
        min = BrassRoomNo
        max = BrassRoomNo + RoomNoOffset
    case String:
        value = &StringNo
        min = StringRoomNo
        max = StringRoomNo + RoomNoOffset
    case Composition:
        value = &CompositionNo
        min = CompositionRoomNo
        max = CompositionRoomNo + RoomNoOffset
    case KTM:
        value = &KtmNo
        min = KtmRoomNo
        max = KtmRoomNo + RoomNoOffset
    default:
        Loge("Unknown category %s", category)
        return nil, 0, 0
    }

    return value, min, max
}