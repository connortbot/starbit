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
}

func NewUDPServer(listener *quic.Listener) *UDPServer {
	return &UDPServer{
		listener: listener,
		ticker:   time.NewTicker(3 * time.Second),
		clients:  make(map[quic.Connection]map[quic.Stream]string),
		done:     make(chan struct{}),
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

// message struct for client-server comms
type Message struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Content  string `json:"content,omitempty"`
}

func (s *UDPServer) handleStream(conn quic.Connection, stream quic.Stream) {
	defer stream.Close()

	buf := make([]byte, 1024)
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
		var msg Message
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
			response := Message{Type: "welcome", Content: "Welcome to the game!"}
			jsonResp, _ := json.Marshal(response)
			stream.Write(jsonResp)
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

			// create update (never include galaxy data)
			update := struct {
				Type        string                `json:"type"`
				PlayerCount int32                 `json:"playerCount"`
				Players     map[string]*pb.Player `json:"players"`
				Started     bool                  `json:"started"`
			}{
				Type:        "tick",
				PlayerCount: s.state.PlayerCount,
				Players:     s.state.Players,
				Started:     s.state.Started,
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

			s.state.mu.Unlock()
			s.mu.Unlock()
		}
	}
}
