package socket

import (
	. "github.com/jinuopti/lilpop-server/log"
	"fmt"
	"github.com/panjf2000/gnet"
)

func OnReadExample(ts *tcpServer, frame []byte, c gnet.Conn) {
	s := string(frame)
	Logd("[%s] id=%d, port=%d, OnReadExample: Local[%s] Remote[%s] Frame[%s]",
		ts.name, ts.id, ts.port, c.LocalAddr(), c.RemoteAddr(), s)

	ts.connectedSockets.Range(func(key, value interface{}) bool {
		addr := key.(string)
		c := value.(gnet.Conn)
		c.AsyncWrite([]byte(fmt.Sprintf("[%s] id=%d, port=%d, heart beating to %s\n", ts.name, ts.id, ts.port, addr)))
		return true
	})
}
