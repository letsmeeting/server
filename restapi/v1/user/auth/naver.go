package auth

import (
	"encoding/json"
	. "github.com/jinuopti/lilpop-server/log"
	"io/ioutil"
	"net/http"
)

const (
	UserInfoNaverAPI = "https://openapi.naver.com/v1/nid/me"
)

type UserInfoNaver struct {
	Resultcode string `json:"resultcode"`
	Message    string `json:"message"`
	Response   struct {
		Email        string `json:"email"`
		Nickname     string `json:"nickname"`
		ProfileImage string `json:"profile_image"`
		Age          string `json:"age"`
		Gender       string `json:"gender"`
		Id           string `json:"id"`
		Name         string `json:"name"`
		Birthday     string `json:"birthday"`
		Birthyear    string `json:"birthyear"`
		Mobile       string `json:"mobile"`
	} `json:"response"`
}

func GetUserInfoNaver(aToken string) (*UserInfoNaver, error) {
	req, err := http.NewRequest("GET", UserInfoNaverAPI, nil)
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
	Logd("NAVER Rsp: [%s]", string(bytes))

	u := &UserInfoNaver{}
	err = json.Unmarshal(bytes, u)
	if err != nil {
		return nil, err
	}
	Logd("Marshal: %v", u)

	return u, nil
}