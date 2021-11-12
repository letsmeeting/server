package user

import (
    "encoding/json"
    "github.com/jinuopti/lilpop-server/restapi/v1/user/auth"
    // "github.com/dgrijalva/jwt-go"
    "github.com/jinuopti/lilpop-server/configure"

    "crypto/sha256"
    "encoding/hex"
    "github.com/jinuopti/lilpop-server/database/gorm/logdb"
    "github.com/jinuopti/lilpop-server/database/gorm/userdb"
    . "github.com/jinuopti/lilpop-server/log"
    . "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
    "net/http"
    "strings"
)

const (
    AuthTypeEmail       = "email"
    AuthTypeGoogle      = "google"
    AuthTypeFacebook    = "facebook"
    AuthTypeKakao       = "kakao"
    AuthTypeNaver       = "naver"
)

type LoginReq struct {
    Id       string         `json:"id" example:"lilpop@lilpop.kr"`
    AuthType string         `json:"auth_type" example:"google"` // google, facebook, kakao, naver, email
    Info     map[string]interface{}    `json:"info"`
}

type loginRsp struct {
    Code    int             `json:"code" example:"0"`
    Message string          `json:"message" example:"success"`
    Token   TokenRsp        `json:"token"`
    Profile Profile         `json:"profile"`
}

type EmailInfo struct {
    Password string         `json:"password"` // EMAIL: password/jwt, SNS: jwt
}

type SnsInfo struct {
    AccessToken         string  `json:"access_token"`
    AccessExpiresIn     int     `json:"access_expires_in"`
    RefreshToken        string  `json:"refresh_token"`
    RefreshExpiresIn    int     `json:"refresh_expires_in"`
}

// LoginHandler
//
// @Summary 사용자 로그인
// @Description 사용자가 로그인을 시도한다
// @Description auth_type "email", "google", "kakao", "naver", "facebook"
// @Tags User
// @Accept json
// @Produce json
// @Param user body LoginReq true "User Login 정보 (auth_type 따라 info 상이)"
// @Success 200 {object} JSONResult{data=loginRsp} "OK"
// @Failure 401 {object} JSONResult{data=string} "ERROR"
// @Failure default {object} JSONResult{data=string} "ERROR"
// @Router /user/login [POST]
func LoginHandler(c echo.Context) error {
    var result JSONResult

    jsonBody := make(map[string]interface{})
    err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
    if err != nil {
        return loginError(c, CodeErrParam, "Json Error: " + err.Error())
    }
    Logd("%s", jsonBody)

    loginUser := LoginReq{
        Id: jsonBody["id"].(string),
        AuthType: strings.ToLower(jsonBody["auth_type"].(string)),
        Info: jsonBody["info"].(map[string]interface{}),
    }

    newUser := true
    user := userdb.FindByIdUser(loginUser.Id)
    if user != nil {
        newUser = false
    }

    // AuthType 마다 인증 분기
    var authResult bool
    switch loginUser.AuthType {
    case AuthTypeEmail:
        // email 로그인의 경우 등록된 user 가 없으면 login 실패 response
        if newUser {
            Logd("not exists user: %s", loginUser.Id)
            result = JSONResult{
                Code:    CodeNotFound,
                Message: "User does not exists",
                Data:    nil,
            }
            return c.JSON(http.StatusUnauthorized, result)
        }
        authResult = authEmail(loginUser, user.Password)
    case AuthTypeGoogle:
        authResult = authGoogle(loginUser, user)
    case AuthTypeFacebook:
        authResult = authFacebook(loginUser, newUser)
    case AuthTypeKakao:
        authResult = authKakao(loginUser, user)
    case AuthTypeNaver:
        authResult = authNaver(loginUser, user)
    }

    // 인증 실패
    if !authResult {
        Logd("Invalid password user: %s", loginUser.Id)
        result = JSONResult{
            Code:    CodeUnauthorized,
            Message: "Authentication failed",
            Data:    nil,
        }
        return c.JSON(http.StatusUnauthorized, result)
    }

    rsp := GetLoginResponse(loginUser.Id, newUser, user)
    if rsp == nil {
        Logd("error, failed to get login response")
        return tokenError(c, CodeError, "Failed to login, invalid user info")
    }

    Logd("Login, Access Token: [%s]", rsp.Token.AccessToken)
    logdb.LogI(logdb.TAG_USER, "Success login user " + loginUser.Id)

    return c.JSON(http.StatusOK, rsp)
}

func GetLoginResponse(userId string, newUser bool, user *userdb.User) *loginRsp {
    config := configure.GetConfig()

    var profile Profile
    if user != nil {
        profile = Profile{
            UserId: user.UserId,
            Email: user.Email,
            Name: user.Name,
            Country: user.Country,
            PhoneNumber: user.PhoneNumber,
            Birthday: user.Birthday,
            Gender: user.Gender,
            Picture: user.Picture,
            CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
            UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
        }
    }

    token, err := UpdateToken(userId)
    if err != nil {
        Logd("error, %s", err)
        return nil
    }

    rsp := &loginRsp{
        Code: CodeOK,
        Message: "Success",
        Token: TokenRsp{
            Result: "success login",
            AccessToken: token.AccessToken,
            AccessExpiresIn: config.Lilpop.JwtAtExpiredTime,
            RefreshToken: token.RefreshToken,
            RefreshExpiresIn: config.Lilpop.JwtRtExpiredTime,
        },
        Profile: profile,
    }
    if newUser {
        rsp.Token.Result = "new user"
    }

    return rsp
}

func authEmail(u LoginReq, dbPassword string) bool {
    pwd := u.Info["password"].(string)

    Logd("password: %s", pwd)

    var password string
    if len(pwd) > 0 {
        password = pwd
    } else {
        return false
    }

    hash := sha256.New()
    hash.Write([]byte(password))
    md := hash.Sum(nil)
    encPwd := hex.EncodeToString(md)
    if encPwd == dbPassword {
        return true
    }

    return false
}

func authSns(u *userdb.User, newUser bool) bool {
    if newUser {
        _ = userdb.InsertUser(u)
    } else {
        _ = userdb.UpdateUser(u)
    }
    return true
}

func authGoogle(u LoginReq, dbUser *userdb.User) bool {
    newUser := true
    if u.Info["access_token"] != nil {
        accessToken := u.Info["access_token"].(string)
        //refreshToken := u.Info["refresh_token"].(string)
        //exp := u.Info["refresh_expires_in"].(float64)
        uInfo, err := auth.GetUserInfoGoogle(accessToken)
        if err != nil {
            Logd("error, %s", err)
            return false
        }
        if dbUser != nil {
            newUser = false
            dbUser.Name = uInfo.Name
            dbUser.Picture = uInfo.Picture
            dbUser.Email = uInfo.Email
            dbUser.AuthType = "google"
        } else {
            dbUser = &userdb.User{
                UserId:   u.Id,
                Name:     uInfo.Name,
                Email:    uInfo.Email,
                Picture:  uInfo.Picture,
                AuthType: "google",
                UserType: "User",
            }
        }
        authSns(dbUser, newUser)
        return true
    }
    Logd("error, access token is nil")
    return false
}

func authFacebook(u LoginReq, newUser bool) bool {
    dbUser := &userdb.User{
        UserId: u.Id,
        Name: u.Id,
        Email: u.Id,
        UserType: "User",
    }
    return authSns(dbUser, newUser)
}

func authKakao(u LoginReq, dbUser *userdb.User) bool {
    newUser := true
    if u.Info["access_token"] != nil {
        uInfo, err := auth.GetUserInfoKakao(u.Info["access_token"].(string))
        if err != nil {
            Loge("error, %s", err)
            return false
        }

        if dbUser != nil {
            newUser = false
            dbUser.Name = uInfo.Properties.Nickname
            dbUser.Picture = uInfo.Properties.ThumbnailImage
            dbUser.Email = uInfo.KakaoAccount.Email
            dbUser.AuthType = "kakao"
        } else {
            dbUser = &userdb.User{
                UserId:   u.Id,
                Name:     uInfo.Properties.Nickname,
                Email:    uInfo.KakaoAccount.Email,
                Picture:  uInfo.Properties.ThumbnailImage,
                AuthType: "kakao",
                UserType: "User",
            }
        }
        authSns(dbUser, newUser)
        return true
    }

    Logd("error, access token is nil")
    return false
}

func authNaver(u LoginReq, dbUser *userdb.User) bool {
    newUser := true
    if u.Info["access_token"] != nil {
        uInfo, err := auth.GetUserInfoNaver(u.Info["access_token"].(string))
        if err != nil {
            Loge("error, %s", err)
            return false
        }

        if dbUser != nil {
            newUser = false
            dbUser.Name = uInfo.Response.Name
            dbUser.Picture = uInfo.Response.ProfileImage
            dbUser.Email = uInfo.Response.Email
            dbUser.Birthday = uInfo.Response.Birthyear
            dbUser.Gender = uInfo.Response.Gender
            dbUser.PhoneNumber = uInfo.Response.Mobile
            dbUser.AuthType = "naver"
            if len(dbUser.Birthday) > 0 {
                dbUser.Birthday = dbUser.Birthday + "-"
            }
            if len(uInfo.Response.Birthday) > 0 {
                dbUser.Birthday = dbUser.Birthday + uInfo.Response.Birthday
            }
        } else {
            dbUser = &userdb.User{
                UserId:   u.Id,
                Name:     uInfo.Response.Name,
                Email:    uInfo.Response.Email,
                Picture:  uInfo.Response.ProfileImage,
                Birthday: uInfo.Response.Birthyear,
                Gender:   uInfo.Response.Gender,
                PhoneNumber: uInfo.Response.Mobile,
                AuthType: "naver",
                UserType: "User",
            }
            if len(dbUser.Birthday) > 0 {
                dbUser.Birthday = dbUser.Birthday + "-"
            }
            if len(uInfo.Response.Birthday) > 0 {
                dbUser.Birthday = dbUser.Birthday + uInfo.Response.Birthday
            }
        }
        authSns(dbUser, newUser)
        return true
    }

    Logd("error, access token is nil")
    return false
}

// registError API 요청 처리 실패  응답
func loginError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
