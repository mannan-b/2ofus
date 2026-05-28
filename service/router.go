package service

import "2ofus/websocket"

// Router adapts the service package to the websocket package's dispatcher interface.
type Router struct{}

func (Router) HandleChat(sender *websocket.Client, data websocket.ChatData) {
	HandleChat(sender, data)
}

func (Router) HandleTyping(sender *websocket.Client, data websocket.TypingData) {
	HandleTyping(sender, data)
}

func (Router) HandleSeen(sender *websocket.Client, data websocket.SeenData) {
	HandleSeen(sender, data)
}
