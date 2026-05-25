package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients []*websocket.Conn

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	clients = append(clients, conn)
	fmt.Println("client connected, total:", len(clients))

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			removeClient(conn)
			fmt.Println("client disconnected, total:", len(clients))
			break
		}

		fmt.Println("got message:", string(msg))

		for _, c := range clients {
			c.WriteMessage(websocket.TextMessage, msg)
		}
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
