package restapi

import (
    "github.com/labstack/echo/v4"
    "net/http"
)

const ApiVersion string = "v1"
const ApiPath = "/api/" + ApiVersion

// Application Return Code
const (
    CodeOK           = 0   // 성공
    CodeError        = 100 // 에러, 기타 모든 에러
    CodeUnknown      = 101 // 알 수 없는 요청
    CodeDuplicate    = 102 // 중복, 이미 존재함
    CodeErrFile      = 103 // 에러, 파일 조작(열기,복사,쓰기) 실패
    CodeUnauthorized = 104 // 인증 실패
    CodeErrDatabase  = 105 // Database 작업 실패
    CodeErrParam     = 106 // Parameter 입력 값 오류
    CodeEmpty        = 107 // 비어있음
    CodeNotFound     = 108 // 찾을 수 없음, 검색실패
)

type JSONResult struct {
    Code    int         `json:"code" example:"0"`
    Message string      `json:"message" example:"Success"`
    Data    interface{} `json:"data"`
}

// SuccessResponse API 요청 처리 성공 응답
func SuccessResponse(c echo.Context, rsp interface{}) error {
    successResult := JSONResult{
        Code:    CodeOK,
        Message: "Success",
        Data:    rsp,
    }
    return c.JSON(http.StatusOK, successResult)
}

// ErrorResponse API 요청 처리 실패  응답
func ErrorResponse(c echo.Context, code int, rsp interface{}) error {
    errResult := JSONResult{
        Code:    code,
        Message: "Error",
        Data:    rsp,
    }
    return c.JSON(http.StatusBadRequest, errResult)
}

func WriteCookie(c echo.Context, cookie *http.Cookie, data interface{}) error {
    c.SetCookie(cookie)
    return c.JSON(http.StatusOK, data)
}
