package chat

import "log"

type HubMessage struct {
	SenderID string
	Payload  []byte
}

type Hub struct {
	clients    map[*Client]bool
	Broadcast  chan *HubMessage
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan *HubMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			log.Printf("[Hub] Client %s registered. Online: %d", client.ID, len(h.clients))

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("[Hub] Client %s unregistered. Online: %d", client.ID, len(h.clients))
			}

		case message := <-h.Broadcast:
			formattedPayload := []byte("[" + message.SenderID + "]: " + string(message.Payload))

			for client := range h.clients {
				if client.ID == message.SenderID {
					continue
				}

				select {
				case client.Send <- formattedPayload:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
