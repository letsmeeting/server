package user

import (
    "github.com/jinuopti/lilpop-server/database/gorm/userdb"
    . "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
    "strconv"
)

// listReq 공유메모리 목록 요청 API
type listReq struct {
    Offset string
    Limit  string
}

// listRsp 공유메모리 목록 응답
type listRsp struct {
    TotalCount int
    Count      int
    Info       []UserListInfo
}

type UserListInfo struct {
    UserId    string
    Name      string
    Email     string
    UserType  string
    CreatedAt string
    UpdatedAt string
}

// ListHandler
//
// @Summary 유저 목록 조회
// @Description 유저 목록 조회 요청 API
// @ID user.list
// @Tags User-Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param offset query string false "List 시작 offset"
// @Param limit query string false "조회할 목록 최대 개수"
// 
// @Success 200 {object} JSONResult{data=listRsp} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /user/list [get]
func ListHandler(c echo.Context) error {
    RequestApi := listReq{
        Offset: c.FormValue("offset"),
        Limit:  c.FormValue("limit"),
    }
    iOffset, err := strconv.Atoi(RequestApi.Offset)
    if err != nil {
        iOffset = 0
    }
    iLimit, err := strconv.Atoi(RequestApi.Limit)
    if err != nil {
        iLimit = 0
    }

    list := userdb.GetUserList(iOffset, iLimit)

    var rsp listRsp
    rsp.TotalCount = int(userdb.GetUserTotalCount())
    rsp.Count = len(list)

    for _, userInfo := range list {
        info := &UserListInfo{
            UserId:     userInfo.UserId,
            Name:       userInfo.Name,
            Email:      userInfo.Email,
            UserType:   userInfo.UserType,
            CreatedAt:  userInfo.CreatedAt.String(),
            UpdatedAt:  userInfo.UpdatedAt.String(),
        }
        rsp.Info = append(rsp.Info, *info)
    }

    return listSuccess(c, rsp)
}

// listSuccess API 요청 처리 성공 응답
func listSuccess(c echo.Context, data listRsp) error {
    return SuccessResponse(c, data)
}

// listError API 요청 처리 실패  응답
func listError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
