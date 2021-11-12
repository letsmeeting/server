package appmmain

import (
	"github.com/jinuopti/lilpop-server/communication/grpc"
	"github.com/jinuopti/lilpop-server/communication/http"
	ws "github.com/jinuopti/lilpop-server/communication/http/websocket"
	. "github.com/jinuopti/lilpop-server/configure"
	"github.com/jinuopti/lilpop-server/database/gorm"
	"github.com/jinuopti/lilpop-server/database/gorm/logdb"
	"github.com/jinuopti/lilpop-server/database/gorm/tokendb"
	"github.com/jinuopti/lilpop-server/database/gorm/userdb"
	"github.com/jinuopti/lilpop-server/database/mongodb"
	. "github.com/jinuopti/lilpop-server/log"
	"sync"
	utility "github.com/jinuopti/lilpop-server/library"
	"github.com/jinuopti/lilpop-server/database/redisdb"
)

func Run(conf *Values) {
	if conf.Lilpop.Enabled == false {
		Logi("Lilpop Application Disabled")
		return
	}
	Logi("Lilpop Application Started")

	wg := new(sync.WaitGroup)

	// Init DATABASE
	err := gormdb.InitSingletonDB()
	if err == nil {
		userdb.InitUserTable()
		logdb.InitLoggingTable()
		tokendb.InitTable()
	}
	if conf.Lilpop.EnableMongoDb {
		err = mongodb.Connect()
		if err != nil {
			Logd("Failed MongoDB connect")
		}
	}
	if conf.Lilpop.EnableRedis {
		_, err = redisdb.InitRedis()
		if err != nil {
			Logd("Failed Redis connect")
		}
	}

	// HTTP
	if conf.Net.EnableHttp && conf.Lilpop.EnableHttpServer {
		go httpserver.HttpServer(conf.Lilpop.HttpServerPort)
	}

	// Websocket
	if conf.Net.EnableWebsocket && conf.Lilpop.EnableWebsocketServer {
		ws.InitWebSocket()
	}

	// gRPC
	if conf.Net.EnableGrpc && conf.Lilpop.EnableGrpcServer {
		go grpcserver.GrpcServer(conf.Lilpop.GrpcServerPort)
	}

	// Start Scheduler
	if conf.Lilpop.SchedulerMinute {
		go utility.SchedulerMin(SchedulerMin)
	}
	if conf.Lilpop.SchedulerHour {
		go utility.SchedulerHour(SchedulerHour)
	}
	if conf.Lilpop.SchedulerDay {
		go utility.SchedulerDay(SchedulerDay)
	}

	wg.Wait()
}
