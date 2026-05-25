package service
import (
	"2OFUS/websocket" // Import your local folder (Module Name / Folder Name)
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
		"from": sender.UserID,
		"to":   data.To,
		"msg":  data.Msg,
	}

	return message
}

func saveMessage(message map[string]interface{}) {

	fmt.Println("saving message to db")
}

func deliverMessage(
	receiver *websocket.Client,
	message map[string]interface{},
) {

	receiver.Conn.WriteJSON(websocket.WSMessage{
		T: "chat",
		D: message,
	})
}

func sendAcknowledgement(sender *websocket.Client) {

	sender.Conn.WriteJSON(websocket.WSMessage{
		T: "ack",
		D: websocket.AckData{
			Status: "sent",
		},
	})
}

func triggerNotification(
	receiver *websocket.Client,
	message map[string]interface{},
) {

	fmt.Println("triggering notification")
}

func HandleChat(sender *websocket.Client, data websocket.ChatData) {
	/*
	1. Validate packet
	2. Identify sender
	3. Create message object
	4. Save to DB
	5. Deliver to recipient
	6. Send acknowledgements
	7. Trigger notifications
	*/
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

	sendAcknowledgement(sender)

	triggerNotification(receiver, message)
}