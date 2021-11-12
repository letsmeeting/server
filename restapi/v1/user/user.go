package user

import (
    "errors"
    . "github.com/jinuopti/lilpop-server/log"
    "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
)

const (
    Admin = "admin"
    User = "user"
    Entertainer = "entertainer"
)

func SetRoute(e *echo.Echo) {
    // jwtConfig := middleware.JWTConfig{
    //     Claims:     &LilpopClaims{},
    //     SigningKey: []byte("secret"),
    // }
    // jwtFunc := middleware.JWTWithConfig(jwtConfig)

    e.POST(restapi.ApiPath+"/user/regist", RegistHandler)
    e.POST(restapi.ApiPath+"/user/token", TokenHandler)
    e.POST(restapi.ApiPath+"/user/login", LoginHandler)
    e.GET(restapi.ApiPath+"/user/profile", ProfileHandler)
    e.POST(restapi.ApiPath+"/user/update", UpdateHandler)
    e.DELETE(restapi.ApiPath+"/user/delete", DeleteHandler)

    // ADMIN
    e.GET(restapi.ApiPath+"/user/list", ListHandler)

    //r := e.Group(restapi.ApiPath + "/user/restricted")
    //r.Use(middleware.JWT([]byte("secret")))
    //r.GET("", Restricted)
}

func CheckHeaderToken(id string, c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if len(token) < (TokenLen + TokenHeaderLen) {
        Logd("Invalid Token len %d, [%s]", len(token), token)
        return errors.New("error, invalid token length")
    }
    tokenHeader := token[:6]
    if tokenHeader != TokenType {
        return errors.New("invalid Bearer type, only \"Bearer ...\"")
    }
    token = token[7:]

    // verify token
    isOk, err := VerifyToken(id, "access_token", token)
    if !isOk || err != nil {
        return err
    }

    return nil
}