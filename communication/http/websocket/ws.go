package ws

import (
    . "github.com/jinuopti/lilpop-server/log"
    "github.com/labstack/echo/v4"
)

var world *World

func InitWebSocket() {
    if world == nil {
        world = NewWorld()
        go world.Run()
    }
    Logd("Initialize WebSocket! uri: /ws")
}

func WebSocketHandler(c echo.Context, readCallback func(*Client, []byte), closeCallback func(*Client)) error {
    conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
    if err != nil {
        return err
    }
    defer func() { _ = conn.Close() }()

    ch := make(chan bool)
    client := NewClient(world, conn, ch, readCallback, closeCallback)
    client.World.ChanEnter <- client

    <- ch

    Logd("Exit WebSocket Handler")

    return nil
}
