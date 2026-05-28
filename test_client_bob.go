package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8080/ws?uid=bob"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer c.Close()

	log.Println("bob connected, waiting for messages...")
	for {
		_, m, err := c.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}
		log.Println("received:", string(m))
		// keep running
		time.Sleep(10 * time.Millisecond)
	}
}
