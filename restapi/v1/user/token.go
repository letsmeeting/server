package user

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/jinuopti/lilpop-server/configure"
	"github.com/jinuopti/lilpop-server/database/gorm/tokendb"
	"github.com/jinuopti/lilpop-server/database/redisdb"
	. "github.com/jinuopti/lilpop-server/log"
	. "github.com/jinuopti/lilpop-server/restapi/v1"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

const (
	TokenHeaderLen = 7
	TokenLen = 143

	TokenType = "Bearer"

	PrefixTokenAccess  = "access_token:"
	PrefixTokenRefresh = "refresh_token:"
)

type LilpopClaims struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	jwt.StandardClaims
}

type TokenReq struct {
	Id 			string `json:"id"`
	GrantType   string `json:"grant_type" example:"access_token or refresh_token"`
	Token 		string `json:"token"`
}

type TokenRsp struct {
	Result 				string	`json:"result"`
	AccessToken			string	`json:"access_token"`
	AccessExpiresIn 	int		`json:"access_expires_in"`
	RefreshToken		string	`json:"refresh_token"`
	RefreshExpiresIn	int		`json:"refresh_expires_in"`
}

// TokenHandler
//
// @Summary Token 만료 확인
// @Description Token 만료 확인 API
// @ID user.token
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param token body TokenReq true "Token 정보"
//
// @Success 200 {object} JSONResult{data=TokenRsp} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /user/token [POST]
func TokenHandler(c echo.Context) error {
	config := configure.GetConfig()

	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {
		return tokenError(c, CodeErrParam, "Json Error: " + err.Error())
	}
	Logd("JSON: %s", jsonBody)

	req := TokenReq{
		Id: jsonBody["id"].(string),
		GrantType: strings.ToLower(jsonBody["grant_type"].(string)),
		Token: jsonBody["token"].(string),
	}

	dbToken := tokendb.FindToken(req.Id)
	if dbToken == nil {     // new_user
		return tokenError(c, CodeNotFound, "Not found user " + req.Id + " token")
	}

	var accessToken string
	var refreshToken string

	switch req.GrantType {
	case "access_token":
		isOk, err := VerifyToken(req.Id, req.GrantType, req.Token)
		if err != nil {
			return tokenError(c, CodeError, err.Error())
		}
		if !isOk {
			return tokenError(c, CodeUnauthorized, "Invalid Token")
		}
		accessToken, refreshToken = IssuedToken(req.Id, "user")
		Logd("access_token: %s", req.Token)
	case "refresh_token":
		isOk, err := VerifyToken(req.Id, req.GrantType, req.Token)
		if err != nil {
			return tokenError(c, CodeError, "Error verify token")
		}
		if !isOk {
			return tokenError(c, CodeUnauthorized, "Invalid Token")
		}
		accessToken, refreshToken = IssuedToken(req.Id, "user")
		Logd("refresh_token: %s", req.Token)

	default:
		Logd("Unknown Grant Type %s", req.GrantType)
		return tokenError(c, CodeUnknown, "Unknown grant_type %s" + req.GrantType)
	}

	now := time.Now().Local()

	accessExpiresIn := now.Add(time.Second * time.Duration(config.Lilpop.JwtAtExpiredTime))
	refreshExpiresIn := now.Add(time.Second * time.Duration(config.Lilpop.JwtRtExpiredTime))

	Logd("AccessExpiresIn[%s], AccessToken[%s]", accessExpiresIn.String(), accessToken)
	Logd("RefreshExpiresIn[%s], RefreshToken[%s]", refreshExpiresIn.String(), refreshToken)

	// redis
	aKey := PrefixTokenAccess + req.Id
	_ = redisdb.Set(aKey, accessToken, time.Second * time.Duration(config.Lilpop.JwtAtExpiredTime))
	rKey := PrefixTokenRefresh + req.Id
	_ = redisdb.Set(rKey, refreshToken, time.Second * time.Duration(config.Lilpop.JwtRtExpiredTime))

	token := &tokendb.Token{
		UserId: req.Id,
		AccessToken: accessToken,
		AccessTokenTime: accessExpiresIn,
		RefreshToken: refreshToken,
		RefreshTokenTime: refreshExpiresIn,
	}
	err = tokendb.UpdateToken(token)
	if err != nil {
		Logd("error, %s", err)
		return tokenError(c, CodeErrDatabase, "Failed insert token to DB")
	}

	tokenRsp := TokenRsp{
		Result: "Valid Token",
		AccessToken: accessToken,
		AccessExpiresIn: config.Lilpop.JwtAtExpiredTime,
		RefreshToken: refreshToken,
		RefreshExpiresIn: config.Lilpop.JwtRtExpiredTime,
	}

	return tokenSuccess(c, tokenRsp)
}

func IssuedToken(id string, userType string) (accessToken string, refreshToken string) {
	config := configure.GetConfig()

	// Create Access Token
	aToken := jwt.New(jwt.SigningMethodHS256)
	// Set claims : Access Token
	aClaims := aToken.Claims.(jwt.MapClaims)
	aClaims["name"] = id
	aClaims["type"] = userType
	aExpired := time.Duration(config.Lilpop.JwtAtExpiredTime)
	aClaims["exp"] = time.Now().Add(time.Second * aExpired).Unix()
	// Generate encoded token and send it as response.
	accessToken, _ = aToken.SignedString([]byte("secret"))

	// Create Refresh Token
	rToken := jwt.New(jwt.SigningMethodHS256)
	// Set claims : Refresh Token
	rClaims := rToken.Claims.(jwt.MapClaims)
	rClaims["name"] = id
	rClaims["type"] = userType
	rExpired := time.Duration(config.Lilpop.JwtRtExpiredTime)
	rClaims["exp"] = time.Now().Add(time.Second * rExpired).Unix()
	// Generate encoded token and send it as response.
	refreshToken, _ = rToken.SignedString([]byte("secret"))

	return accessToken, refreshToken
}

func UpdateToken(userId string) (*tokendb.Token, error) {
	var err error
	config := configure.GetConfig()

	accessToken, refreshToken := IssuedToken(userId, "user")
	now := time.Now().Local()
	accessExpiresIn := now.Add(time.Second * time.Duration(config.Lilpop.JwtAtExpiredTime))
	refreshExpiresIn := now.Add(time.Second * time.Duration(config.Lilpop.JwtRtExpiredTime))
	token := &tokendb.Token{
		UserId: userId,
		AccessToken: accessToken,
		AccessTokenTime: accessExpiresIn,
		RefreshToken: refreshToken,
		RefreshTokenTime: refreshExpiresIn,
	}

	// redis
	aKey := PrefixTokenAccess + userId
	_ = redisdb.Set(aKey, accessToken, time.Second * time.Duration(config.Lilpop.JwtAtExpiredTime))
	rKey := PrefixTokenRefresh + userId
	_ = redisdb.Set(rKey, refreshToken, time.Second * time.Duration(config.Lilpop.JwtRtExpiredTime))

	find := tokendb.FindToken(userId)
	if find != nil {
		err = tokendb.UpdateToken(token)
	} else {
		err = tokendb.InsertToken(token)
	}
	if err != nil {
		return nil, err
	}

	return token, nil
}

func VerifyToken(id string, grantType string, token string) (bool, error) {
	//user := c.Get("user").(*jwt.Token)
	//claims := user.Claims.(*LilpopClaims)

	conf := configure.GetConfig()
	now := time.Now()
	dbToken := tokendb.FindToken(id)
	if dbToken == nil {
		Logd("id=%s, dbToken is nil, token:[%s]", id, token)
		return true, nil
	}

	aKey := PrefixTokenAccess + id
	aToken, err := redisdb.GetString(aKey)
	if err == redis.Nil {
		Logd("REDIS Key:[%s] AccessToken is nil", aKey)
	} else if err != nil {
		Logd("REDIS error, %s", err)
	} else {
		Logd("REDIS AccessToken Key:[%s], Token=[%s]", aKey, aToken)
	}
	rKey := PrefixTokenRefresh + id
	rToken, err := redisdb.GetString(rKey)
	if err == redis.Nil {
		Logd("REDIS Key:[%s] RefreshToken is nil", rKey)
	} else if err != nil {
		Logd("REDIS error, %s", err)
	} else {
		Logd("REDIS RefreshToken Key:[%s], Token=[%s]", rKey, rToken)
	}

	if grantType == "access_token" {
		diff := now.Sub(dbToken.AccessTokenTime)
		if diff.Seconds() <= float64(conf.Lilpop.JwtAtExpiredTime) {
			if dbToken.AccessToken == token {
				Logd("Access Token is valid [%s]", token)
				return true, nil
			} else {
				Logd("Access Token is invalid [%s]", token)
				return false, errors.New("No match access token")
			}
		} else {
			Logd("Access Token expired")
			return false, errors.New("Access Token expired")
		}
	} else if grantType == "refresh_token" {
		diff := now.Sub(dbToken.RefreshTokenTime)
		if diff.Seconds() <= float64(conf.Lilpop.JwtRtExpiredTime) {
			if dbToken.RefreshToken == token {
				Logd("Refresh Token is valid [%s]", token)
				return true, nil
			} else {
				Logd("Refresh Token is invalid [%s]", token)
				return false, errors.New("No match refresh token")
			}
		} else {
			Logd("Refresh Token expired")
			return false, errors.New("Refresh Token expired")
		}
	} else {
		return false, errors.New("Unknown Grant Type")
	}
}

// tokenSuccess API 요청 처리 성공 응답
func tokenSuccess(c echo.Context, data TokenRsp) error {
	return SuccessResponse(c, data)
}

// updateError API 요청 처리 실패  응답
func tokenError(c echo.Context, code int, message string) error {
	return ErrorResponse(c, code, message)
}
