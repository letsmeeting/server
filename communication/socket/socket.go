package socket

import (
	posUtil "github.com/jinuopti/lilpop-server/library"
	"github.com/panjf2000/gnet"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	connCnt int64
)

type tcpServer struct {
	id	 int		// server id (1 ~ increment)
	port int		// server listening port
	name string		// connection name

	*gnet.EventServer
	tick             time.Duration
	connectedSockets sync.Map
}

func newTcpServer(name string, port int) *tcpServer {
	atomic.AddInt64(&connCnt, 1)
	return &tcpServer{id: int(connCnt), port: port, name: name, tick: time.Second}
}

func TcpServerRun(name string, portString string) {
	portList := posUtil.ParseTokenInt(portString, ",")
	for _, p := range portList {
		protoAddr := "tcp://:" + strconv.Itoa(p)
		ts := newTcpServer(name, p)
		go gnet.Serve(ts, protoAddr, gnet.WithMulticore(true), gnet.WithReusePort(true), gnet.WithTicker(true))
	}
}
