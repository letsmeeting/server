package log

import (
    "github.com/jinuopti/lilpop-server/restapi/v1"
    "github.com/labstack/echo/v4"
)

func SetRoute(e *echo.Echo) {
    e.GET(restapi.ApiPath+"/log/find", FindHandler)
}
