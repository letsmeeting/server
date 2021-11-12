package user

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "github.com/jinuopti/lilpop-server/database/gorm/logdb"
    "github.com/jinuopti/lilpop-server/database/gorm/userdb"
    . "github.com/jinuopti/lilpop-server/log"
    . "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
    "net/http"
)

// registReq 유저 회원가입 요청 API
type registReq struct {
    UserId      string `json:"user_id" example:"user"`
    Email       string `json:"email" example:"user@mail.com"`
    Password    string `json:"password" example:"password"`
    Name        string `json:"name" example:"Kim"`
    Country     string `json:"country" example:"Korea, Republic of"` // South Korea, America ...
    PhoneNumber string `json:"phone_number" example:"01011112222"` // 010xxxxyyyy
    Birthday    string `json:"birthday" example:"2000-01-01"`
    Introduce   string `json:"introduce" example:"I am a genius"`
    UserType    string `json:"user_type" example:"User"` // Entertainer or User
    Category    string `json:"category" example:"Singer"` // Singer,Actor,Comedian,Dancer,Magician,Etc
}

// RegistHandler
//
// @Summary 유저 회원가입
// @Description 유저 회원가입 요청 API
// @ID user.regist
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param user body registReq true "User Info json"
//
// @Success 200 {object} JSONResult{data=loginRsp} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /user/regist [post]
func RegistHandler(c echo.Context) error {
    jsonBody := make(map[string]interface{})
    err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
    if err != nil {
        return registError(c, CodeErrParam, "Json Error: " + err.Error())
    }

    Logd("%s", jsonBody)

    id := jsonBody["user_id"].(string)
    // 중복 체크
    dbUser := userdb.FindByIdUser(id)
    if dbUser != nil {
        return registError(c, CodeDuplicate, "Already registered ID: " + id)
    }

    hash := sha256.New()
    hash.Write([]byte(jsonBody["password"].(string)))
    md := hash.Sum(nil)
    mdStr := hex.EncodeToString(md)

    //var birthday time.Time
    //if len(jsonBody["birthday"].(string)) == 10 {
    //    layOut := "2000-01-01"
    //    birthday, err = time.Parse(layOut, jsonBody["birthday"].(string))
    //}

    newUser := userdb.User{
        UserId: id,
        Email: jsonBody["email"].(string),
        Password: mdStr,
        Name: jsonBody["name"].(string),
        Country: jsonBody["country"].(string),
        PhoneNumber: jsonBody["phone_number"].(string),
        Birthday: jsonBody["birthday"].(string),
        Introduce: jsonBody["introduce"].(string),
        UserType: jsonBody["user_type"].(string),
        Category: jsonBody["category"].(string),
    }
    _ = userdb.InsertUser(&newUser)

    rsp := GetLoginResponse(id, true, dbUser)
    if rsp == nil {
        Logd("error, GetLoginResponse rsp is nil")
        return registError(c, CodeError, "Invalid user info " + id)
    }
    rsp.Token.Result = "success regist"

    // Write to Log DB
    logdb.WriteLog(&logdb.LogModel{
        LogType: "USER",
        Severity: logdb.Info,
        Message: "Completed the new user registration : " + id,
    })

    return c.JSON(http.StatusOK, rsp)
}

// registError API 요청 처리 실패  응답
func registError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
