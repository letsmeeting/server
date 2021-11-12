package log

import (
    . "github.com/jinuopti/lilpop-server/log"
    . "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
)

const MaxFindRecord = 1000

type FindRequest struct {
    Name        string
    Offset      string
    Limit       string
    Severity    string
    Start       string
    End         string
    Keyword     string
}

type FindResponse struct {
    TotalCount int         // 조회한 결과 총 개수
    Count      int         // 응답할 결과 개수 (Limit)
    Logs       []ResultLog // 로그 내용 slice
}

type ResultLog struct {
    Id        int
    Timestamp string
    LogType   string
    Severity  string
    Message   string
}

// FindHandler
//
// @Summary Log 검색 요청
// @Description Log 검색 요청 API
// @Description - 시간검색 및 문자열 Query
// @Description - severity : bit or 로 중복선택 (bit 0:ERROR, bit 1:INFO, bit 2:WARNING, bit 4:DEBUG)
// @Description   예) 1=ERROR, 3=ERROR & INFO, 6=INFO & WARNING, 8=DEBUG, 15=ALL
// @Description - 기본정렬 : 최신이 가장 위로 (desc)
// @Tags Log
// @Accept json
// @Produce json
// @Security ApiKeyAuth
//
// @Param name query string true "App 이름 (쉼표(,)로 복수개 선택가능 예: engine,ai,hmi)"
// @Param offset query string false "검색 시작 offset"
// @Param limit query string true "검색 결과 limit"
// @Param severity query string false "검색 로그 종류 (기본 ALL, bit or - 상단 description 참조)"
// @Param start query string false "시간검색 시작시간 (형식: 2021-06-01 12:30:00)"
// @Param end query string false "시간검색 종료시간 (형식: 2021-12-31 23:59:59)"
// @Param keyword query string false "검색할 문자열 Keyword"
//
// @Success 200 {object} JSONResult{data=string} "Success"
// @Failure 400 {object} JSONResult{data=string} "Error"
// @Failure 401 {object} JSONResult{data=string} "Unauthorized"
// @Router /log/find [get]
func FindHandler(c echo.Context) error {
    var req FindRequest
    req.Name = c.FormValue("name")
    req.Offset = c.FormValue("offset")
    req.Limit = c.FormValue("limit")
    req.Severity = c.FormValue("severity")
    req.Start = c.FormValue("start")
    req.End = c.FormValue("end")
    req.Keyword = c.FormValue("keyword")

    Logd("Request: Name[%s], Offset[%s], Limit[%s], Severity[%s], Start[%s], End[%s], Keyword[%s]",
        req.Name, req.Offset, req.Limit, req.Severity, req.Start, req.End, req.Keyword)

    return findSuccess(c, "success")
}

// findSuccess API 요청 처리 성공 응답
func findSuccess(c echo.Context, r string) error {
    return SuccessResponse(c, r)
}

// findError API 요청 처리 실패  응답
func findError(c echo.Context, code int, message string) error {
    return ErrorResponse(c, code, message)
}
