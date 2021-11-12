package auth

import (
	"encoding/json"
	. "github.com/jinuopti/lilpop-server/log"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	KakaoUserInfoAPI = "https://kapi.kakao.com/v2/user/me"
)

type UserInfoKakao struct {
	Id          int       `json:"id"`
	ConnectedAt time.Time `json:"connected_at"`
	Properties  struct {
		Nickname       string `json:"nickname"`
		ProfileImage   string `json:"profile_image"`
		ThumbnailImage string `json:"thumbnail_image"`
	} `json:"properties"`
	KakaoAccount struct {
		ProfileNicknameNeedsAgreement bool `json:"profile_nickname_needs_agreement"`
		ProfileImageNeedsAgreement    bool `json:"profile_image_needs_agreement"`
		Profile                       struct {
			Nickname          string `json:"nickname"`
			ThumbnailImageUrl string `json:"thumbnail_image_url"`
			ProfileImageUrl   string `json:"profile_image_url"`
			IsDefaultImage    bool   `json:"is_default_image"`
		} `json:"profile"`
		HasEmail            bool   `json:"has_email"`
		EmailNeedsAgreement bool   `json:"email_needs_agreement"`
		IsEmailValid        bool   `json:"is_email_valid"`
		IsEmailVerified     bool   `json:"is_email_verified"`
		Email               string `json:"email"`
	} `json:"kakao_account"`
}

func GetUserInfoKakao(aToken string) (*UserInfoKakao, error) {
	req, err := http.NewRequest("GET", KakaoUserInfoAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer " + aToken)

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	// 결과 출력
	bytes, _ := ioutil.ReadAll(rsp.Body)
	Logd("KAKAO Rsp: [%s]", string(bytes))

	u := &UserInfoKakao{}
	err = json.Unmarshal(bytes, u)
	if err != nil {
		return nil, err
	}
	Logd("Marshal: %v", u)

	return u, nil
}