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

	default:
		fmt.Println("unknown event:", msg.T)
	}
}
