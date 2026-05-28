package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8080/ws?uid=alice"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer c.Close()

	msg := map[string]interface{}{
		"t": "chat",
		"d": map[string]interface{}{
			"id":  "test-mid-1",
			"to":  "bob",
			"msg": "Hello from alice (automated test)",
		},
	}

	payload, _ := json.Marshal(msg)

	if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
		log.Fatal("write error:", err)
	}

	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, m, err := c.ReadMessage()
	if err != nil {
		log.Println("read error (likely no immediate response):", err)
		return
	}

	log.Println("received:", string(m))
}
