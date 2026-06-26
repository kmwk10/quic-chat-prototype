package chat

import (
	"context"
	"io"
	"log"

	"github.com/quic-go/webtransport-go"
)

type Client struct {
	ID      string
	Hub     *Hub
	Session *webtransport.Session
	Send    chan []byte
}

func NewClient(id string, hub *Hub, session *webtransport.Session) *Client {
	return &Client{
		ID:      id,
		Hub:     hub,
		Session: session,
		Send:    make(chan []byte, 256),
	}
}

func (c *Client) WritePump() {
	defer func() {
		_ = c.Session.CloseWithError(0, "write pump closed")
		c.Hub.Unregister <- c
	}()

	for msg := range c.Send {
		stream, err := c.Session.OpenUniStream()
		if err != nil {
			log.Printf("[Client %s] Failed to open UniStream: %v", c.ID, err)
			return
		}

		_, err = stream.Write(msg)
		if err != nil {
			log.Printf("[Client %s] Failed to write to stream: %v", c.ID, err)
			_ = stream.Close()
			return
		}
		_ = stream.Close()
	}
}

func (c *Client) ReadPump() {
	defer func() {
		_ = c.Session.CloseWithError(0, "read pump closed")
		c.Hub.Unregister <- c
	}()

	for {
		stream, err := c.Session.AcceptStream(context.Background())
		if err != nil {
			log.Printf("[Client %s] Session closed or AcceptStream error: %v", c.ID, err)
			break
		}

		go func(stream *webtransport.Stream) {
			defer stream.Close()

			msg, err := io.ReadAll(stream)
			if err != nil {
				log.Printf("[Client %s] Failed to read from stream: %v", c.ID, err)
				return
			}

			if len(msg) == 0 {
				return
			}

			log.Printf("[Client %s] Message received: %s", c.ID, string(msg))

			c.Hub.Broadcast <- &HubMessage{
				SenderID: c.ID,
				Payload:  msg,
			}
		}(stream)
	}
}
