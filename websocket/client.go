package websocket

import "github.com/gorilla/websocket"

type Client struct {
	UserID string
	Conn   *websocket.Conn
}

var Clients = map[string]*Client{}