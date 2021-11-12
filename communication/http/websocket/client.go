package ws

import (
    "github.com/gorilla/websocket"
    utility "github.com/jinuopti/lilpop-server/library"
    . "github.com/jinuopti/lilpop-server/log"
    "time"
)

const (
    // Time allowed to write a message to the peer.
    writeWait = 10 * time.Second
    // Time allowed to read the next pong message from the peer.
    pongWait = 60 * time.Second
    // Send pings to peer with this period. Must be less than pongWait.
    pingPeriod = pongWait - writeWait
    // Maximum message size allowed from peer.
    maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  maxMessageSize,
    WriteBufferSize: maxMessageSize,
}

type Client struct{
    World *World
    Conn *websocket.Conn
    Send chan []byte
    Exit chan bool
    Disconnect bool

    Login  bool
    UserId string

    ReadCallback func(*Client, []byte)
    CloseCallback func(*Client)
}

type RequestJson struct {
    Cmd     string
    Data    interface{}
}

func NewClient (w *World, c *websocket.Conn, e chan bool, f func(*Client, []byte), cf func(*Client)) (client *Client) {
    client = &Client{
        World: w,
        Conn: c,
        Send: make(chan []byte, maxMessageSize),
        Exit: e,
        ReadCallback: f,
        CloseCallback: cf,
    }
    go client.read()
    go client.write()

    return client
}

func (c *Client) Close() {
    c.World.ChanLeave <- c
    _ = c.Conn.Close()
}

func (c *Client) read() {
    defer func() {
        c.World.ChanLeave <- c
        _ = c.Conn.Close()
    }()

    c.Conn.SetReadLimit(maxMessageSize)
    _ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
    c.Conn.SetPongHandler(func(string) error { _ = c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                Logd("%v", err)
            }
            break
        }

        if c.ReadCallback != nil {
            c.ReadCallback(c, message)
        } else {
            c.Send <- message
        }

        //c.world.broadcast <- message
    }
    c.Exit <- true
}

func (c *Client) write() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        _ = c.Conn.Close()
    }()
    for {
        select {
        case message, ok := <-c.Send:
            _ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                _ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            prettyJson, err := utility.GetPrettyJsonStr(message)
            if err == nil {
                Logd("Write: \n%s", prettyJson)
            }
            _ = c.Conn.WriteMessage(websocket.TextMessage, message)
            if c.Disconnect {
                c.Close()
            }
        case tick := <-ticker.C:
            // Logd("Send PingMessage, pingPeriod: %s, message: %s", pingPeriod.String(), tick.String())
            _ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, []byte(tick.String())); err != nil { return }
        }
    }
}
