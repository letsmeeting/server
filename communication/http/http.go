package httpserver

import (
	"github.com/jinuopti/lilpop-server/application/lilpop"
	"github.com/jinuopti/lilpop-server/restapi/v1"
	ApiLog "github.com/jinuopti/lilpop-server/restapi/v1/log"
	ApiUser "github.com/jinuopti/lilpop-server/restapi/v1/user"

	"net/http"
	"strings"

	"github.com/jinuopti/lilpop-server/configure"
	. "github.com/jinuopti/lilpop-server/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jinuopti/lilpop-server/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// HttpServer
// @title Lilpop REST API
// @version 0.1.0
// @BasePath /api/v1
// @query.collection.format multi
//
// @description <h2><b>Lilpop REST API Swagger Documentation</b></h2>
// @description
// @description <b>Application Response Code</b>
// @Description <b>CodeOK</b> (0) : 성공
// @Description <b>CodeError</b> (100) : 에러, 기타 모든 에러
// @Description <b>CodeUnknown</b> (101) : 알 수 없는 요청
// @Description <b>CodeDuplicate</b> (102) : 중복, 이미 존재함
// @Description <b>CodeErrFile</b> (103) : 에러, 파일 조작(열기,복사,쓰기) 실패
// @Description <b>CodeUnauthorized</b> (104) : 인증 실패
// @Description <b>CodeErrDatabase</b> (105) : Database 작업 실패
// @Description <b>CodeErrParam</b> (106) : Parameter 입력 값 오류
// @Description <b>CodeEmpty</b> (107) : 비어있음
// @Description <b>CodeNotFound</b> (108) : 찾을 수 없음, 검색실패
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func HttpServer(port string) {
	conf := configure.GetConfig()

	port = strings.Trim(port, " ")
	if len(port) <= 0 {
		port = conf.Lilpop.HttpServerPort
	}
	if len(conf.Lilpop.SwaggerPort) > 0 {
		docs.SwaggerInfo.Host = conf.Lilpop.HttpListenAddr + ":" + conf.Lilpop.SwaggerPort
	} else {
		docs.SwaggerInfo.Host = conf.Lilpop.HttpListenAddr
	}

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Logger.SetOutput(GetLogWriter())

	// websocket
	if conf.Net.EnableWebsocket && conf.Lilpop.EnableWebsocketServer {
		Logd("Websocket Enabled")
		e.GET("/ws", lilpop.WsAppHandler)
	}
	// swagger documents
	if conf.Test.EnableSwagger {
		Logd("Swagger Enabled")
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
	// Health check
	e.GET("/health", healthHandler)

	setRoute(e)

	Logd("HTTP ListenAddr: %s, Port: %s", conf.Lilpop.HttpListenAddr, port)

	if conf.Lilpop.EnableSSL {
		certFile := conf.Lilpop.SslCertFile
		keyFile := conf.Lilpop.SslKeyFile
		if err := e.StartTLS(":" + port, certFile, keyFile); err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	} else {
		err := e.Start(":" + port)
		if err != nil {
			e.Logger.Fatal(err)
		}
	}
}

// setRoute
func setRoute(e *echo.Echo) {
	// Insert API Route
	ApiUser.SetRoute(e)
	ApiLog.SetRoute(e)
}

func healthHandler(c echo.Context) error {
	// Logd("calling health check API")
	return restapi.SuccessResponse(c, nil)
}