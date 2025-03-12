package game

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"time"

	pb "starbit/proto"

	"github.com/quic-go/quic-go"
)

type UDPTickMsg string

type UDPClient struct {
	session  quic.Connection
	stream   quic.Stream
	username string
	tickCh   chan UDPTickMsg
}

// message struct for client-server communication
type Message struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Content  string `json:"content,omitempty"`
}

// game update from server via UDP
type GameUpdate struct {
	Type        string                `json:"type"`
	PlayerCount int32                 `json:"playerCount"`
	Players     map[string]*pb.Player `json:"players"`
	Started     bool                  `json:"started"`
	Galaxy      *pb.GalaxyState       `json:"galaxy,omitempty"`
}

func NewUDPClient() *UDPClient {
	return &UDPClient{
		tickCh: make(chan UDPTickMsg, 10),
	}
}

func (c *UDPClient) Connect() error {
	// create a context with timeout for the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect to the server
	session, err := quic.DialAddr(ctx, "localhost:50052", &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"starbit-quic"},
	}, nil)

	if err != nil {
		log.Printf("UDP connection error: %v", err)
		return err
	}

	c.session = session
	log.Println("UDP connection established successfully")

	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		log.Printf("Failed to open stream: %v", err)
		return err
	}

	c.stream = stream
	go c.handleStream(stream)
	return nil
}

func (c *UDPClient) Register(username string) error {
	c.username = username

	// create registration message
	msg := Message{
		Type:     "register",
		Username: username,
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// send registration message
	_, err = c.stream.Write(jsonMsg)
	if err != nil {
		log.Printf("Failed to send registration: %v", err)
		return err
	}

	return nil
}

func (c *UDPClient) GetTickChannel() <-chan UDPTickMsg {
	return c.tickCh
}

func (c *UDPClient) handleStream(stream quic.Stream) {
	defer stream.Close()
	buf := make([]byte, 4096) // larger buffer for game state

	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Println("Stream read error:", err)
			return
		}

		// try to parse as a game update
		var gameUpdate GameUpdate
		if err := json.Unmarshal(buf[:n], &gameUpdate); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// handle different message types
		switch gameUpdate.Type {
		case "tick":
			c.tickCh <- UDPTickMsg("UDP Tick Received")

		case "welcome":
			log.Println("Registered with UDP server successfully")
		}
	}
}
