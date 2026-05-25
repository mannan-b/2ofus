package main

package main

import (
	"fmt"
	"log"
	"net/http"
	"2OFUS/websocket"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWs(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	userID := r.URL.Query().Get("uid")

	client := &websocket.Client{
		UserID: userID,
		Conn:   conn,
	}

	websocket.Clients[userID] = client

	fmt.Println("client connected:", userID)

	for {
		var msg websocket.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			delete(websocket.Clients, userID)
			fmt.Println("client disconnected:", userID)
			break
		}
		websocket.HandleMessage(client, msg)
	}
}

func removeClient(conn *websocket.Conn) {
	for i, c := range clients {
		if c == conn {
			clients = append(clients[:i], clients[i+1:]...)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", handleWs)

	fmt.Println("server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
