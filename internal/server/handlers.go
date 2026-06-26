package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kmwk10/quic-chat-prototype/internal/chat"
	"github.com/quic-go/webtransport-go"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to write health response: %v", err)
	}
}

func HandleWebTransport(hub *chat.Hub, wtServer *webtransport.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := wtServer.Upgrade(w, r)
		if err != nil {
			log.Printf("[Server] Failed to upgrade to WebTransport: %v", err)
			return
		}

		b := make([]byte, 4)
		_, _ = rand.Read(b)
		clientID := fmt.Sprintf("user_%x", b)

		client := chat.NewClient(clientID, hub, session)
		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
