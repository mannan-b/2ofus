package websocket
import "encoding/json"


//event wrapper
type WSMessage struct {
	T string          `json:"t"`
	D json.RawMessage `json:"d"`
}

//chat-type data
type ChatData struct {
	To      string `json:"to"`
	Msg string `json:"msg"`
	ID string `json:"id"`
}

//typing-indicator
type TypingData struct {
	To string `json:"to"`
}

//seen acknowledgment MID-messageID
type SeenData struct {
	MID string `json:"mid"`
}

//online-offline status UID-userid, S-status
type PresenceData struct {
	UID string `json:"uid"`
	Online bool   `json:"online"`
}

//notification message Tt-title, Bd-body
type NotificationData struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

//acknowledgement schema
type AckData struct {
	Status string `json:"status"`
}