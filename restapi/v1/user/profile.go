package user

import (
    "github.com/jinuopti/lilpop-server/database/gorm/logdb"
    "github.com/jinuopti/lilpop-server/database/gorm/userdb"
    . "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
)

type Profile struct {
    UserId      string `json:"user_id" example:"user"`
    Email       string `json:"email" example:"user@mail.com"`
    Name        string `json:"name" example:"홍길동"`
    Country     string `json:"country" example:"KR"` // South Korea, America ...
    PhoneNumber string `json:"phone_number" example:"01011112222"` // 010xxxxyyyy
    Picture     string `json:"picture" example:"https://picture.link.address"`
    Gender      string `json:"gender" example:"{male or female}"`
    Birthday    string `json:"birthday" example:"2000-01-01"`
    Category    string `json:"category" example:"음악"`

    CreatedAt   string
    UpdatedAt   string
}

// ProfileHandler
//
// @Summary 유저 조회
// @Description 유저 조회 요청 API
// @ID user.profile
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param id query string true "조회할 User Id"
//
// @Success 200 {object} JSONResult{data=Profile} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /user/profile [get]
func ProfileHandler(c echo.Context) error {
    id := c.FormValue("id")

    //err := CheckHeaderToken(id, c)
    //if err != nil {
    //    return profileError(c, CodeUnauthorized, err.Error())
    //}

    user := userdb.FindByIdUser(id)
    if user == nil {
        return profileError(c, CodeUnauthorized, "Username does not exist")
    }

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

    logdb.LogD(logdb.TAG_USER, "Request profile user " + id)

    return profileSuccess(c, profile)
}

// profileSuccess API 요청 처리 성공 응답
func profileSuccess(c echo.Context, data Profile) error {
    return SuccessResponse(c, data)
}

// profileError API 요청 처리 실패  응답
func profileError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
