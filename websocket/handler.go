package websocket

import (
	"encoding/json"
	"fmt"
)

// Dispatcher lets the websocket package route messages without importing service directly.
type Dispatcher interface {
	HandleChat(*Client, ChatData)
	HandleTyping(*Client, TypingData)
	HandleSeen(*Client, SeenData)
}

func HandleMessage(client *Client, msg WSMessage, dispatcher Dispatcher) {
	switch msg.T {

	case "chat":
		var data ChatData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("chat parse error:", err)
			return
		}
		dispatcher.HandleChat(client, data)

	case "typing":
		var data TypingData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("typing parse error:", err)
			return
		}
		dispatcher.HandleTyping(client, data)

	case "seen":
		var data SeenData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("seen parse error:", err)
			return
		}
		dispatcher.HandleSeen(client, data)

	case "key":
		// Relay public key to the intended recipient. Server does not store private material.
		var payload struct {
			To  string `json:"to"`
			Pub string `json:"pub"`
		}

		if err := json.Unmarshal(msg.D, &payload); err != nil {
			fmt.Println("key parse error:", err)
			return
		}

		fmt.Println("relaying key from", client.UserID, "to", payload.To)

		receiver, ok := Clients[payload.To]
		if !ok {
			fmt.Println("key relay: receiver not connected:", payload.To)
			return
		}

		// forward with sender metadata
		forward := map[string]string{"from": client.UserID, "pub": payload.Pub}
		b, _ := json.Marshal(forward)
		receiver.Conn.WriteJSON(WSMessage{T: "key", D: b})

	default:
		fmt.Println("unknown event:", msg.T)
	}
}
