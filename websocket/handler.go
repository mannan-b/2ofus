package websocket

import (
	"encoding/json"
	"fmt"
	"2OFUS/service"

)

func HandleMessage(client *Client,msg WSMessage) {

	switch msg.T {

	case "chat":
		var data ChatData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("chat parse error:", err)
			return
		}

		handleChat(data)

	case "typing":
		var data TypingData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("typing parse error:", err)
			return
		}

		handleTyping(data)

	case "seen":
		var data SeenData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("seen parse error:", err)
			return
		}

		handleSeen(data)

	case "presence":
		var data PresenceData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("presence parse error:", err)
			return
		}

		handlePresence(data)

	case "notification":
		var data NotificationData

		err := json.Unmarshal(msg.D, &data)
		if err != nil {
			fmt.Println("notification parse error:", err)
			return
		}

		handleNotification(data)

	default:
		fmt.Println("unknown event:", msg.T)
	}
}