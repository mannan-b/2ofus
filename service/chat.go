package service
import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"2ofus/websocket"
)

func validateChat(data websocket.ChatData) bool {

	if data.To == "" {
		return false
	}

	if data.Msg == "" {
		return false
	}

	if len(data.Msg) > 1000 {
		return false
	}

	return true
}

func getReceiver(userID string) (*websocket.Client, bool) {

	receiver, ok := websocket.Clients[userID]

	return receiver, ok
}

func handleOfflineUser(userID string) {

	fmt.Println("user offline:", userID)
}

func createMessage(
	sender *websocket.Client,
	data websocket.ChatData,
) map[string]interface{} {

	message := map[string]interface{}{
		"id":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"from": sender.UserID,
		"to":   data.To,
		"msg":  data.Msg,
	}

	return message
}

func saveMessage(message map[string]interface{}) {

	file, _ := os.OpenFile(
		"messages.json",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	defer file.Close()

	bytes, _ := json.Marshal(message)

	file.Write(bytes)
	file.Write([]byte("\n"))
}

func deliverMessage(
	receiver *websocket.Client,
	message map[string]interface{},
) {

	receiver.Conn.WriteJSON(websocket.WSMessage{
		T: "chat",
		D: mustMarshal(message),
	})
}

func sendAcknowledgement(
	sender *websocket.Client,
	message map[string]interface{},
) {

	ack := map[string]interface{}{
		"id":     message["id"],
		"status": "sent",
	}

	sender.Conn.WriteJSON(websocket.WSMessage{
		T: "ack",
		D: mustMarshal(ack),
	})
}
func triggerNotification(
	receiver *websocket.Client,
	message map[string]interface{},
) {

	fmt.Println("triggering notification")
}

func HandleChat(sender *websocket.Client, data websocket.ChatData) {

	if !validateChat(data) {
		fmt.Println("invalid chat packet")
		return
	}

	receiver, ok := getReceiver(data.To)

	if !ok {
		handleOfflineUser(data.To)
		return
	}

	message := createMessage(sender, data)

	saveMessage(message)

	deliverMessage(receiver, message)

	sendAcknowledgement(sender, message)
	triggerNotification(receiver, message)
}

func HandleTyping(sender *websocket.Client, data websocket.TypingData) {

	receiver, ok := getReceiver(data.To)

	if !ok {
		return
	}

	payload := map[string]interface{}{
		"from": sender.UserID,
	}

	receiver.Conn.WriteJSON(websocket.WSMessage{
		T: "typing",
		D: mustMarshal(payload),
	})
}

func HandleSeen(sender *websocket.Client, data websocket.SeenData) {

	for _, client := range websocket.Clients {

		client.Conn.WriteJSON(websocket.WSMessage{
			T: "seen",
			D: mustMarshal(map[string]interface{}{
				"mid": data.MID,
			}),
		})
	}
}

func mustMarshal(v interface{}) json.RawMessage {

	bytes, err := json.Marshal(v)

	if err != nil {
		return nil
	}

	return bytes
}