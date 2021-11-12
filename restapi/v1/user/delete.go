package user

import (
    "github.com/jinuopti/lilpop-server/database/gorm/userdb"
    . "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
)

// deleteReq 유저 Delete 요청 API
type deleteReq struct {
    UserId string
}

// deleteRsp 유저 Delete 응답 API
type deleteRsp struct {
    Result string
}

// DeleteHandler
//
// @Summary 유저 Delete
// @Description 유저 Delete 요청 API
// @ID user.delete
// @Tags User-Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param id query string true "삭제할 User Id"
//
// @Success 200 {object} JSONResult{data=deleteRsp} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /user/delete [DELETE]
func DeleteHandler(c echo.Context) error {
    id := c.FormValue("id")

    user := userdb.FindByIdUser(id)
    if user == nil {
        return deleteError(c, CodeUnauthorized, "Username does not exist")
    }

    err := userdb.DeleteUser(id)

    if err != nil {
        return deleteError(c, CodeErrDatabase, "Failed delete user from database")
    }

    // Write to Log DB
    //logdb.WriteLog(&logdb.Lilpop{
    //    LogType: "Engine",
    //    Severity: logdb.Info,
    //    Message: "Deleted User Info Id : " + id,
    //})

    return deleteSuccess(c, "Success Delete User Info")
}

// deleteSuccess API 요청 처리 성공 응답
func deleteSuccess(c echo.Context, message string) error {
    return SuccessResponse(c, message)
}

// deleteError API 요청 처리 실패  응답
func deleteError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
