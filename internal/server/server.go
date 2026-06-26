package server

import (
	"crypto/tls"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/kmwk10/quic-chat-prototype/internal/chat"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

type Server struct {
	addr     string
	certFile string
	keyFile  string
	wtServer *webtransport.Server
}

func New(addr, certFile, keyFile string, hub *chat.Hub) *Server {
	r := chi.NewRouter()

	h3Server := &http3.Server{
		Addr:            addr,
		Handler:         r,
		EnableDatagrams: true,
		TLSConfig:       http3.ConfigureTLSConfig(&tls.Config{}),
	}

	wtServer := &webtransport.Server{H3: h3Server}
	webtransport.ConfigureHTTP3Server(h3Server)

	r.Get("/health", HandleHealth)
	r.HandleFunc("/wt", HandleWebTransport(hub, wtServer))

	return &Server{
		addr:     addr,
		certFile: certFile,
		keyFile:  keyFile,
		wtServer: wtServer,
	}
}

func (s *Server) Start() error {
	log.Printf("[Server] Starting HTTP/3 + WebTransport server on %s...", s.addr)
	return s.wtServer.ListenAndServeTLS(s.certFile, s.keyFile)
}
