package lilpop

const (
	// App -> Server
	CmdLogin = "login"			// Websocket 로그인
	CmdStart = "start"			// 매칭 시작
	CmdExit = "exit"			// 매칭 종료 (대기실로 퇴장)
	CmdTrickle = "trickle"		// ICE candidate
	CmdNext = "next"			// 다음 매칭, 지금 방을 퇴장하고 다음 매칭 대기를 요청

	// Server -> App
	CmdJoined = "joined"		// 방 입장
	CmdRequest = "request"		// 요청 (offer, ...)
	CmdReady = "ready"			// 방 입장 및 매칭 준비 완료 (상대 유저 입장 대기)
	CmdComplete = "complete"	// 영상대화를 위한 모든 준비 완료
	CmdSuccess = "success"		// 각종 App 요청에 대한 성공 응답
	CmdFail = "fail"			// 각종 App 요청에 대한 실패 응답
	CmdAck = "ack"				// App으로 부터 일반적인 데이터 수신 시 응답 (trickle, ...)

	// Common
	CmdOffer = "offer"			// offer SDP 송신
	CmdAnswer = "answer"		// answer SDP 송신
	CmdEvent = "event"			// 대화 중 각종 event 송신 (video, audio, hangup, ...)
)

type SimpleReq struct {
	Lilpop			string	`json:"lilpop" example:"{command}"`
}

type GeneralReq struct {
	Lilpop 			string `json:"lilpop" example:"{command}"`
	Transaction 	string `json:"transaction" example:"{12byte random string}"`
}

type GeneralRsp struct {
	Lilpop 			string `json:"lilpop" example:"{success or fail}"`
	Transaction 	string `json:"transaction" example:"{12byte random string}"`
	Message 		string `json:"message" example:"{message}"`
}

// LoginReq 로그인 요청
type LoginReq struct {
	Lilpop 			string 		`json:"lilpop" example:"login"`
	Transaction 	string 		`json:"transaction" example:"YdJwaD03jdUY"`
	Data			LoginBody 	`json:"data"`
}
type LoginBody struct {
	UserId 		string		`json:"user_id" example:"lilpop@lilpop.kr"`
	AccessToken	string		`json:"access_token" example:"{access_token}"`
}

// LoginRsp 로그인 응답
type LoginRsp struct {
	Lilpop 			string 	`json:"lilpop" example:"login"`
	Transaction 	string 	`json:"transaction" example:"YdJwaD03jdUY"`
}

// StartReq 매칭시작 요청
type StartReq struct {
	Lilpop 			string		`json:"lilpop" example:"start"`
	Transaction 	string		`json:"transaction" example:"YdJwaD03jdUY"`
	Data	 		StartBody	`json:"data"`
}
type StartBody struct {
	Category 		string	`json:"category" example:"보컬,드러머,댄서"`
	Region 			string	`json:"region" example:"대한민국"`
	Tag 			string 	`json:"tag" example:"친해지기좋은시간,일단와요,기타치고,노래하고"`
}

// JoinedRsp 매칭시작 응답
type JoinedRsp struct {
	Lilpop			string	`json:"lilpop" example:"joined"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
	Publishers 		int 	`json:"publishers" example:"0"`
}

// RequestReq 서버에서 App 으로 요청
type RequestReq struct {
	Lilpop 		string	`json:"lilpop" example:"request"`
	Type 		string	`json:"type" example:"offer"`
}

// MatchStopReq 매칭종료 요청
type MatchStopReq struct {
	Lilpop 			string	`json:"lilpop" example:"match_stop"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
}

// MatchStopRsp 매칭종료 응답
type MatchStopRsp struct {
	Lilpop 			string 	`json:"lilpop" example:"match_stop"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
}

// MatchJoinEvent 사용자 입장
type MatchJoinEvent struct {
	Lilpop 			string 	`json:"lilpop" example:"match_ready"`
	UserId 			string 	`json:"user_id" example:"{user id}"`
	AudioCodec 		string 	`json:"audio_codec" example:"opus"`
	VideoCodec 		string 	`json:"video_codec" example:"vp8"`
}

// SdpApp App 에서 서버로 offer or answer 전송
type SdpApp struct {
	Lilpop			string	`json:"lilpop" example:"{offer or answer}"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
	Data			SdpAppData `json:"data"`
}
type SdpAppData struct {
	Audio 		bool	`json:"audio,omitempty" example:"true"`
	Video 		bool 	`json:"video,omitempty" example:"true"`
	Sdp 		SdpBody	`json:"sdp"`
}
type SdpBody struct {
	Type        string  `json:"type" example:"{offer or answer}"`
	Description string  `json:"description"`
}

// SdpServerOffer 서버에서 App 으로 offer or answer 전송
type SdpServerOffer struct {
	Lilpop			string	`json:"lilpop" example:"offer"`
	UserId       	string	`json:"user_id" example:"tom"`
	Data			SdpServerData `json:"data"`
}
type SdpServerAnswer struct {
	Lilpop			string	`json:"lilpop" example:"answer"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
	Data			SdpServerData `json:"data"`
}
type SdpServerData struct {
	AudioCodec 		string	`json:"audio_codec" example:"opus"`
	VideoCodec		string	`json:"video_codec" example:"vp8"`
	Sdp 			string	`json:"sdp"`
}

type TrickleData struct {
	Lilpop			string	`json:"lilpop" example:"trickle"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
	Data 			map[string]interface{} 	`json:"data"`
}

type EventServer struct {
	Lilpop			string	`json:"lilpop" example:"event"`
	UserId 			string	`json:"user_id"`
	Type 			string	`json:"type"`
	Value 			string 	`json:"value"`
}

type EventApp struct {
	Lilpop			string	`json:"lilpop" example:"event"`
	Transaction 	string	`json:"transaction" example:"YdJwaD03jdUY"`
	Type 			string	`json:"type"`
	Value 			string 	`json:"value"`
}

