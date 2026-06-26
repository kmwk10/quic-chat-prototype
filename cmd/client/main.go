package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	log.Println("[Test 1] Checking HTTP/3 health endpoint...")
	h3Transport := &http3.Transport{TLSClientConfig: tlsConfig}
	defer h3Transport.Close()

	h3Client := &http.Client{Transport: h3Transport}
	resp, err := h3Client.Get("https://127.0.0.1:4433/health")
	if err != nil {
		log.Fatalf("[Client] HTTP/3 request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("[Client] Protocol: %s, Response: %s", resp.Proto, string(body))
	log.Println(strings.Repeat("-", 40))

	log.Println("[Test 2] Initializing WebTransport connection...")
	dialer := webtransport.Dialer{
		TLSClientConfig: tlsConfig,
		QUICConfig: &quic.Config{
			EnableDatagrams:                  true,
			EnableStreamResetPartialDelivery: true,
			KeepAlivePeriod:                  10 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, session, err := dialer.Dial(ctx, "https://127.0.0.1:4433/wt", nil)
	if err != nil {
		log.Fatalf("[Client] Failed to connect via WebTransport: %v", err)
	}
	defer session.CloseWithError(0, "client exit")
	log.Println("[Client] Successfully connected via WebTransport!")

	go func() {
		for {
			stream, err := session.AcceptUniStream(context.Background())
			if err != nil {
				log.Printf("\n[Client] Server connection closed: %v", err)
				return
			}

			message, err := io.ReadAll(stream)
			if err != nil {
				log.Printf("\n[Client] Failed to read message from stream: %v", err)
				stream.CancelRead(0)
				continue
			}

			fmt.Printf("%s\n> ", string(message))
		}
	}()

	fmt.Println("--- Welcome to QUIC Chat! Type /exit to quit ---")
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			fmt.Print("> ")
			continue
		}
		if text == "/exit" {
			log.Println("[Client] Exiting chat...")
			break
		}

		stream, err := session.OpenStream()
		if err != nil {
			log.Printf("[Client] Failed to open stream: %v", err)
			break
		}
		_, err = stream.Write([]byte(text))
		stream.Close()
		if err != nil {
			log.Printf("[Client] Failed to write to network: %v", err)
			break
		}
		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[Client] Error reading input: %v", err)
	}
}
