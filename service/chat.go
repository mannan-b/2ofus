package service

import (
	"2ofus/websocket"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var messagesMu sync.Mutex

func validateChat(data websocket.ChatData) bool {

	if data.To == "" {
		return false
	}

	// allow either plaintext msg or encrypted ciphertext
	if data.Msg == "" && data.Ciphertext == "" {
		return false
	}

	if data.Msg != "" && len(data.Msg) > 1000 {
		return false
	}

	if data.Ciphertext != "" && len(data.Ciphertext) > 20000 {
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

	id := fmt.Sprintf("%d", time.Now().UnixNano())

	// If ciphertext is present, preserve encrypted fields.
	if data.Ciphertext != "" {
		message := map[string]interface{}{
			"id":   id,
			"from": sender.UserID,
			"to":   data.To,
			"ct":   data.Ciphertext,
			"nonce": data.Nonce,
			"spub": data.SenderPub,
		}
		return message
	}

	message := map[string]interface{}{
		"id":   id,
		"from": sender.UserID,
		"to":   data.To,
		"msg":  data.Msg,
	}

	return message
}

// saveMessage writes the message to two places:
// - per-recipient file messages_<recipient>.json
// - websocket_messages.json (global log of relayed encrypted/plain messages)
func saveMessage(message map[string]interface{}) {
	// marshal once
	bytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println("saveMessage: marshal error:", err)
		return
	}

	cwd, _ := os.Getwd()

	// write per-recipient file
	to, _ := message["to"].(string)
	if to == "" {
		to = "unknown"
	}
	perPath := filepath.Join(cwd, fmt.Sprintf("messages_%s.json", to))

	messagesMu.Lock()
	defer messagesMu.Unlock()

	if err := appendJSONLine(perPath, bytes); err != nil {
		fmt.Println("saveMessage: per-recipient write error:", err)
	} else {
		fmt.Println("saved per-recipient message to", perPath)
	}

	// append to websocket global log
	wsPath := filepath.Join(cwd, "websocket_messages.json")
	if err := appendJSONLine(wsPath, bytes); err != nil {
		fmt.Println("saveMessage: websocket log write error:", err)
	} else {
		fmt.Println("saved websocket message to", wsPath)
	}
}

func appendJSONLine(path string, bytes []byte) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		return err
	}
	if _, err := file.Write([]byte("\n")); err != nil {
		return err
	}
	return file.Sync()
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
