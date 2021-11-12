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

// updateReq 유저 회원가입 요청 API
type updateReq struct {
    Email       string `json:"email" example:"user@mail.com"`
    Password    string `json:"password" example:"password"`
    Name        string `json:"name" example:"홍길동"`
    Gender      string `json:"gender" example:"male"`
    Country     string `json:"country" example:"KR"` // South Korea, America ...
    PhoneNumber string `json:"phone_number" example:"01011112222"` // 010xxxxyyyy
    Birthday    string `json:"birthday" example:"2000-01-01"`
    Category    string `json:"category" example:"음악"` // Singer,Actor,Comedian,Dancer,Magician,Etc
}

type updateRsp struct {
    Code    int     `json:"code" example:"0"`
    Message string  `json:"message" example:"success"`
    Profile Profile `json:"data"`
}

// UpdateHandler
//
// @Summary 유저 정보 수정
// @Description 유저 정보 수정 요청 API
// @ID user.edit
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param id query string true "Update User Id"
// @Param token body updateReq true "User 수정 정보"
//
// @Success 200 {object} JSONResult{data=Profile} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /user/update [POST]
func UpdateHandler(c echo.Context) error {
    id := c.FormValue("id")
    if len(id) > 0 {
        err := CheckHeaderToken(id, c)
        if err != nil {
            return registError(c, CodeUnauthorized, err.Error())
        }
    }

    j := make(map[string]interface{})
    err := json.NewDecoder(c.Request().Body).Decode(&j)
    if err != nil {
        return registError(c, CodeErrParam, "Json Error: " + err.Error())
    }

    Logd("%s", j)

    if len(id) <= 0 {
        if j["user_id"] != nil {
            id = j["user_id"].(string)
        } else {
            return registError(c, CodeErrParam, "user_id is empty")
        }
    }

    user := userdb.FindByIdUser(id)
    if user == nil {
        return registError(c, CodeError, "user does not exists " + id)
    }

    var mdStr string
    if j["password"] != nil {
        password := j["password"].(string)
        hash := sha256.New()
        hash.Write([]byte(password))
        md := hash.Sum(nil)
        if password == "" {
            mdStr = ""
        } else {
            mdStr = hex.EncodeToString(md)
        }
        user.Password = mdStr
    }

    if j["name"] != nil {
        user.Name = j["name"].(string)
    }
    if j["email"] != nil {
        user.Email = j["email"].(string)
    }
    if j["country"] != nil {
        user.Country = j["country"].(string)
    }
    if j["phone_number"] != nil {
        user.PhoneNumber = j["phone_number"].(string)
    }
    if j["gender"] != nil {
        user.Gender = j["gender"].(string)
    }
    if j["birthday"] != nil {
        user.Birthday = j["birthday"].(string)
    }
    if j["category"] != nil {
        user.Category =  j["category"].(string)
    }

    _ = userdb.UpdateUser(user)

    profile := Profile{
        UserId:      user.UserId,
        Email:       user.Email,
        Name:        user.Name,
        Country:     user.Country,
        PhoneNumber: user.PhoneNumber,
        Birthday:    user.Birthday,
        Picture:     user.Picture,
        Gender:      user.Gender,
        Category:    user.Category,
        CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
    }

    rsp := &updateRsp{
        Code: CodeOK,
        Message: "success update user information",
        Profile: profile,
    }

    logdb.LogI(logdb.TAG_USER, "Update user info " + rsp.Profile.UserId)

    return c.JSON(http.StatusOK, rsp)
}

// updateError API 요청 처리 실패  응답
func updateError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
