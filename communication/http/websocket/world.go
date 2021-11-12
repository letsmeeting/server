package ws

import (
    . "github.com/jinuopti/lilpop-server/log"
    //"time"
)

type World struct {
    clientMap map[*Client]bool
    ChanEnter chan *Client
    ChanLeave chan *Client
    broadcast chan []byte
}

func NewWorld() *World {
    return &World{
        clientMap: make(map[*Client]bool, 5),
        broadcast: make(chan []byte),
    }
}

func (w *World) Run() {
    w.ChanEnter = make(chan *Client)
    w.ChanLeave = make(chan *Client)

    //ticker := time.NewTicker(10 * time.Second)

    for {
        select {
        case client := <- w.ChanEnter:
            w.clientMap[client] = true
            Logd("New WebSocket Client, Len: %d", len(w.clientMap))
        case client := <- w.ChanLeave:
            if _, ok := w.clientMap[client]; ok {
                if client.CloseCallback != nil {
                    client.CloseCallback(client)
                }
                delete(w.clientMap, client)
                close(client.Send)
            }
            Logd("Exit WebSocket Client, Len: %d", len(w.clientMap))
        case message := <- w.broadcast:
            for client := range w.clientMap {
                client.Send <- message
            }
        // case tick := <- ticker.C:
        //     for client := range w.clientMap {
        //         client.send <- []byte(tick.String())
        //     }
        }
    }
}