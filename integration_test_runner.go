package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"2ofus/service"
	"2ofus/websocket"
	gws "github.com/gorilla/websocket"
)

func main() {
	cleanupFiles()
	websocket.Clients = map[string]*websocket.Client{}

	server := startTestServer()
	defer server.Close()

	aliceConn := dialClient(server.URL, "alice")
	defer aliceConn.Close()

	bobConn := dialClient(server.URL, "bob")
	defer bobConn.Close()

	curve := ecdh.X25519()
	alicePriv, err := curve.GenerateKey(rand.Reader)
	must(err)
	bobPriv, err := curve.GenerateKey(rand.Reader)
	must(err)

	alicePubB64 := base64.StdEncoding.EncodeToString(alicePriv.PublicKey().Bytes())
	bobPubB64 := base64.StdEncoding.EncodeToString(bobPriv.PublicKey().Bytes())

	alicePubCh := make(chan []byte, 1)
	bobPubCh := make(chan []byte, 1)
	chatCh := make(chan map[string]string, 1)
	ackCh := make(chan struct{}, 1)

	go readLoop("alice", aliceConn, alicePubCh, ackCh, nil, "")
	go readLoop("bob", bobConn, bobPubCh, nil, chatCh, bobPubB64)

	must(aliceConn.WriteJSON(map[string]interface{}{
		"t": "key",
		"d": map[string]string{"to": "bob", "pub": alicePubB64},
	}))

	bobPub := waitBytes(bobPubCh, 3*time.Second, "bob public key")
	alicePub := waitBytes(alicePubCh, 3*time.Second, "alice public key")

	alicePeerKey, err := curve.NewPublicKey(alicePub)
	must(err)
	bobPeerKey, err := curve.NewPublicKey(bobPub)
	must(err)

	// alice uses bob's public key; bob uses alice's public key.
	aliceShared, err := alicePriv.ECDH(alicePeerKey)
	must(err)
	bobShared, err := bobPriv.ECDH(bobPeerKey)
	must(err)
	if !bytesEqual(aliceShared, bobShared) {
		panic("shared secrets differ")
	}

	sharedKey := sha256.Sum256(aliceShared)
	plaintext := []byte("Hello from alice (automated test)")
	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	must(err)

	block, err := aes.NewCipher(sharedKey[:])
	must(err)
	gcm, err := cipher.NewGCM(block)
	must(err)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	must(aliceConn.WriteJSON(map[string]interface{}{
		"t": "chat",
		"d": map[string]string{
			"id":    "runner-mid-enc-1",
			"to":    "bob",
			"ct":    base64.StdEncoding.EncodeToString(ciphertext),
			"nonce": base64.StdEncoding.EncodeToString(nonce),
			"spub":  alicePubB64,
		},
	}))

	select {
	case <-ackCh:
		fmt.Println("alice got ack")
	case <-time.After(3 * time.Second):
		panic("timeout waiting for alice ack")
	}

	bobChat := waitChat(chatCh, 5*time.Second)
	gotCiphertext, err := base64.StdEncoding.DecodeString(bobChat["ct"])
	must(err)
	gotNonce, err := base64.StdEncoding.DecodeString(bobChat["nonce"])
	must(err)

	bobBlock, err := aes.NewCipher(sharedKey[:])
	must(err)
	bobGCM, err := cipher.NewGCM(bobBlock)
	must(err)
	decrypted, err := bobGCM.Open(nil, gotNonce, gotCiphertext, nil)
	must(err)
	if string(decrypted) != string(plaintext) {
		panic(fmt.Sprintf("decrypted text mismatch: %q", string(decrypted)))
	}

	must(waitForFileContains(filepath.Join(".", "messages_bob.json"), "\"ct\"", 5*time.Second))
	must(waitForFileContains(filepath.Join(".", "websocket_messages.json"), "\"ct\"", 5*time.Second))

	fmt.Println("SUCCESS: shared secret matched, encrypted chat decrypted, and messages were saved")
	_ = bobPub
}

func startTestServer() *httptest.Server {
	upgrader := gws.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrade error:", err)
			return
		}

		userID := r.URL.Query().Get("uid")
		if userID == "" {
			userID = "anonymous"
		}

		client := &websocket.Client{UserID: userID, Conn: conn}
		websocket.Clients[userID] = client
		fmt.Println("client connected:", userID)

		defer func() {
			delete(websocket.Clients, userID)
			_ = conn.Close()
			fmt.Println("client disconnected:", userID)
		}()

		for {
			var msg websocket.WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			fmt.Println("received ws message type:", msg.T)
			websocket.HandleMessage(client, msg, service.Router{})
		}
	})

	return httptest.NewServer(mux)
}

func dialClient(serverURL, userID string) *gws.Conn {
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws?uid=" + userID
	conn, _, err := gws.DefaultDialer.Dial(wsURL, nil)
	must(err)
	return conn
}

func readLoop(name string, conn *gws.Conn, pubCh chan<- []byte, ackCh chan<- struct{}, chatCh chan<- map[string]string, replyPubB64 string) {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(name, "read error:", err)
			return
		}

		var pkt struct {
			T string          `json:"t"`
			D json.RawMessage `json:"d"`
		}
		if err := json.Unmarshal(raw, &pkt); err != nil {
			fmt.Println(name, "packet error:", err)
			continue
		}

		switch pkt.T {
		case "ack":
			if ackCh != nil {
				ackCh <- struct{}{}
			}
		case "key":
			var d struct {
				From string `json:"from"`
				Pub  string `json:"pub"`
			}
			if err := json.Unmarshal(pkt.D, &d); err != nil {
				fmt.Println(name, "key parse error:", err)
				continue
			}

			pub, err := base64.StdEncoding.DecodeString(d.Pub)
			if err != nil {
				fmt.Println(name, "pub decode error:", err)
				continue
			}

			if pubCh != nil {
				pubCh <- pub
			}

			if name == "bob" && replyPubB64 != "" {
				// Bob replies with his own public key once he receives Alice's public key.
				reply := map[string]interface{}{
					"t": "key",
					"d": map[string]string{"to": d.From, "pub": replyPubB64},
				}
				must(conn.WriteJSON(reply))
			}
		case "chat":
			var d map[string]string
			if err := json.Unmarshal(pkt.D, &d); err != nil {
				fmt.Println(name, "chat parse error:", err)
				continue
			}
			if chatCh != nil {
				chatCh <- d
			}
		}
	}
}

func waitBytes(ch <-chan []byte, timeout time.Duration, label string) []byte {
	select {
	case b := <-ch:
		return b
	case <-time.After(timeout):
		panic("timeout waiting for " + label)
	}
}

func waitChat(ch <-chan map[string]string, timeout time.Duration) map[string]string {
	select {
	case d := <-ch:
		return d
	case <-time.After(timeout):
		panic("timeout waiting for chat")
	}
}

func waitForFileContains(path, substr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(b), substr) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s to contain %q", path, substr)
}

func cleanupFiles() {
	_ = os.Remove("messages_bob.json")
	_ = os.Remove("websocket_messages.json")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
