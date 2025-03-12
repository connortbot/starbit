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

type ErrorMessage struct {
	Content  string
	Username string
}

type UDPClient struct {
	session  quic.Connection
	stream   quic.Stream
	username string
	tickCh   chan *pb.TickMsg
	errorCh  chan ErrorMessage
}

type ServerMessage struct {
	Type     string      `json:"type"`
	Username string      `json:"username,omitempty"`
	Content  string      `json:"content,omitempty"`
	TickMsg  *pb.TickMsg `json:"tickMsg,omitempty"`
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
		tickCh:  make(chan *pb.TickMsg, 10),
		errorCh: make(chan ErrorMessage, 10),
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
	msg := ServerMessage{
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

func (c *UDPClient) GetTickChannel() <-chan *pb.TickMsg {
	return c.tickCh
}

func (c *UDPClient) GetErrorChannel() <-chan ErrorMessage {
	return c.errorCh
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

		var serverMsg ServerMessage
		if err := json.Unmarshal(buf[:n], &serverMsg); err != nil {
			log.Printf("Failed to parse as ServerMessage: %v", err)
		}

		switch serverMsg.Type {
		case "tick":
			if serverMsg.TickMsg != nil {
				c.tickCh <- serverMsg.TickMsg
			} else {
				tickMsg := &pb.TickMsg{
					Message: "UDP Tick Received",
				}
				c.tickCh <- tickMsg
			}
		case "welcome":
			log.Println("Registered with UDP server successfully")
		case "error":
			log.Printf("Error from server: %s", serverMsg.Content)
			c.errorCh <- ErrorMessage{
				Content:  serverMsg.Content,
				Username: serverMsg.Username,
			}
		}
	}
}

func (c *UDPClient) SendFleetCreation(fleetCreation *pb.FleetCreation) error {
	tickMsg := &pb.TickMsg{
		Message:        "fleet_creation",
		FleetCreations: []*pb.FleetCreation{fleetCreation},
	}

	msg := ServerMessage{
		Type:     "fleet_creation",
		Username: c.username,
		TickMsg:  tickMsg,
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = c.stream.Write(jsonMsg)
	if err != nil {
		return err
	}

	return nil
}

func (c *UDPClient) SendFleetMovement(fleetMovement *pb.FleetMovement) error {
	tickMsg := &pb.TickMsg{
		Message:        "fleet_movement",
		FleetMovements: []*pb.FleetMovement{fleetMovement},
	}

	msg := ServerMessage{
		Type:     "fleet_movement",
		Username: c.username,
		TickMsg:  tickMsg,
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = c.stream.Write(jsonMsg)
	if err != nil {
		log.Printf("Failed to send fleet movement: %v", err)
		return err
	}

	return nil
}
