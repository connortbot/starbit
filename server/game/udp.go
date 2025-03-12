package game

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	pb "starbit/proto"

	"github.com/quic-go/quic-go"
)

type UDPServer struct {
	listener *quic.Listener
	state    *State
	ticker   *time.Ticker
	clients  map[quic.Connection]map[quic.Stream]string // map conn->streams->username
	mu       sync.Mutex
	done     chan struct{}

	tickMessage *pb.TickMsg
}

func NewUDPServer(listener *quic.Listener) *UDPServer {
	return &UDPServer{
		listener:    listener,
		ticker:      time.NewTicker(3 * time.Second),
		clients:     make(map[quic.Connection]map[quic.Stream]string),
		done:        make(chan struct{}),
		tickMessage: &pb.TickMsg{},
	}
}

// shared state setter
func (s *UDPServer) SetState(state *State) {
	s.state = state
}

func (s *UDPServer) Start() {
	go s.broadcastTicks()

	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			log.Println("UDP accept error:", err)
			continue
		}
		s.mu.Lock()
		s.clients[conn] = make(map[quic.Stream]string)
		s.mu.Unlock()
		go s.handleConn(conn)
	}
}

func (s *UDPServer) handleConn(conn quic.Connection) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Println("Stream accept error:", err)
			return
		}

		go s.handleStream(conn, stream)
	}
}

type ServerMessage struct {
	Type     string      `json:"type"`
	Username string      `json:"username,omitempty"`
	Content  string      `json:"content,omitempty"`
	TickMsg  *pb.TickMsg `json:"tickMsg,omitempty"`
}

func (s *UDPServer) handleStream(conn quic.Connection, stream quic.Stream) {
	defer stream.Close()

	buf := make([]byte, 2048)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Println("Stream read error:", err)
			s.mu.Lock()
			delete(s.clients[conn], stream)
			s.mu.Unlock()
			return
		}

		// decode message
		var msg ServerMessage
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			log.Printf("Message decode error: %v", err)
			continue
		}

		// handle message types
		switch msg.Type {
		case "register":
			s.mu.Lock()
			s.clients[conn][stream] = msg.Username
			log.Printf("UDP client registered: %s", msg.Username)
			s.mu.Unlock()

			// send welcome message
			response := ServerMessage{Type: "welcome", Content: "Welcome to the game!"}
			jsonResp, _ := json.Marshal(response)
			stream.Write(jsonResp)

		case "fleet_movement":
			if msg.TickMsg != nil && len(msg.TickMsg.FleetMovements) > 0 {
				err := s.state.MoveFleet(msg.Username, msg.TickMsg.FleetMovements[0])
				if err != nil {
					log.Printf("Error moving fleet: %v", err)

					errorMsg := ServerMessage{
						Type:     "error",
						Content:  err.Error(),
						Username: msg.Username,
					}

					jsonError, _ := json.Marshal(errorMsg)
					stream.Write(jsonError)
				} else {
					s.mu.Lock()
					s.tickMessage.FleetMovements = append(s.tickMessage.FleetMovements, msg.TickMsg.FleetMovements[0])
					s.mu.Unlock()
				}
			}
		}
	}
}

// broadcastTicks sends game state updates to all connected clients
func (s *UDPServer) broadcastTicks() {
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			if s.state == nil {
				continue
			}

			s.mu.Lock()
			s.state.mu.Lock()

			update := ServerMessage{
				Type:    "tick",
				TickMsg: s.tickMessage,
			}
			jsonUpdate, _ := json.Marshal(update)

			// send to all connected clients
			for _, streams := range s.clients {
				for stream := range streams {
					if _, err := stream.Write(jsonUpdate); err != nil {
						log.Printf("Error sending tick: %v", err)
					}
				}
			}
			s.tickMessage = &pb.TickMsg{}

			s.state.mu.Unlock()
			s.mu.Unlock()
		}
	}
}
