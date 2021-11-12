package auth

import (
	"google.golang.org/api/option"

	"context"

	firebase "firebase.google.com/go/v4"

	"github.com/jinuopti/lilpop-server/configure"
	. "github.com/jinuopti/lilpop-server/log"
)

type UserInfoGoogle struct {
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
}

func GetUserInfoGoogle(aToken string) (*UserInfoGoogle, error) {
	conf := configure.GetConfig()

	opt := option.WithCredentialsFile(conf.Lilpop.FirebaseCredentialFile)
	config := &firebase.Config{ProjectID: conf.Lilpop.FirebaseProjectId}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return nil, err
	}
	auth, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	t, err := auth.VerifyIDToken(context.Background(), aToken)
	if err != nil {
		return nil, err
	}
	Logd("Claim: [%v]", t)

	u := &UserInfoGoogle{}
	if t.Claims["name"] != nil {
		u.Name = t.Claims["name"].(string)
	}
	if t.Claims["picture"] != nil {
		u.Picture = t.Claims["picture"].(string)
	}
	if t.Claims["email"] != nil {
		u.Email = t.Claims["email"].(string)
	}

	return u, nil
}
