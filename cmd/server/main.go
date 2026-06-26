package main

import (
	"log"

	"github.com/kmwk10/quic-chat-prototype/internal/chat"
	"github.com/kmwk10/quic-chat-prototype/internal/server"
)

func main() {
	hub := chat.NewHub()
	go hub.Run()

	srv := server.New(":4433", "cert.pem", "key.pem", hub)

	if err := srv.Start(); err != nil {
		log.Fatalf("[Server] Fatal error: %v", err)
	}
}
