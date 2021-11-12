package socket

import (
	. "github.com/jinuopti/lilpop-server/log"
	"fmt"
	"github.com/panjf2000/gnet"
	"time"
)

// OnInitComplete fires when the server is ready for accepting connections.
// The parameter:server has information and various utilities.
func (ts *tcpServer) OnInitComplete(svr gnet.Server) (action gnet.Action) {
	Logd("[%s] id=%d, port=%d, OnInitComplete callback", ts.name, ts.id, ts.port)
	return
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (ts *tcpServer) OnShutdown(svr gnet.Server) {
	Logd("[%s] id=%d, port=%d, OnShutdown callback", ts.name, ts.id, ts.port)
}

// OnOpened fires when a new connection has been opened.
// The parameter:c has information about the connection such as it's local and remote address.
// Parameter:out is the return value which is going to be sent back to the client.
func (ts *tcpServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	Logd("[%s] id=%d, port=%d, [%s] OnOpened from [%s]", ts.name, ts.id, ts.port, c.LocalAddr(), c.RemoteAddr())
	ts.connectedSockets.Store(c.RemoteAddr().String(), c)
	return
}

// OnClosed fires when a connection has been closed.
// The parameter:err is the last known connection error.
func (ts *tcpServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	Logd("[%s] id=%d, port=%d, [%s] OnClosed from [%s]", ts.name, ts.id, ts.port, c.LocalAddr(), c.RemoteAddr())
	ts.connectedSockets.Delete(c.RemoteAddr().String())
	return
}

// PreWrite fires just before any data is written to any client socket, this event function is usually used to
// put some code of logging/counting/reporting or any prepositive operations before writing data to client.
func (ts *tcpServer) PreWrite() {
	Logd("[%s] id=%d, port=%d, PreWrite callback", ts.name, ts.id, ts.port)
}

// React fires when a connection sends the server data.
// Call c.Read() or c.ReadN(n) within the parameter:c to read incoming data from client.
// Parameter:out is the return value which is going to be sent back to the client.
func (ts *tcpServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	switch ts.name {
	case "example":
		OnReadExample(ts, frame, c)
	default:
		Loge("Unknown connection id=%d, port=%d, name:[%s]", ts.id, ts.port, ts.name)
	}
	return
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (ts *tcpServer) Tick() (delay time.Duration, action gnet.Action) {
	ts.connectedSockets.Range(func(key, value interface{}) bool {
		addr := key.(string)
		c := value.(gnet.Conn)
		c.AsyncWrite([]byte(fmt.Sprintf("[%s] id=%d, port=%d, heart beating to %s\n", ts.name, ts.id, ts.port, addr)))
		return true
	})
	delay = ts.tick
	//Logd("[%s] id=%d, port=%d, Tick callback, delay=%d", ts.name, ts.id, ts.port, delay)
	return
}
