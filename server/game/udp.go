package game

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	pb "starbit/proto"
	galaxy "starbit/server/game/galaxy"

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
		ticker:      time.NewTicker(1 * time.Second),
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
				ownerChange, err := s.state.MoveFleet(msg.Username, msg.TickMsg.FleetMovements[0])
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

					if ownerChange != nil {
						s.tickMessage.SystemOwnerChanges = append(s.tickMessage.SystemOwnerChanges, ownerChange)
					}

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

			remainingBattlingSystems := []int32{}
			for _, systemId := range s.state.battlingSystems {
				// Process battles
				battleActive, healthUpdates, fleetsDestroyed, newOwner := galaxy.ExecuteBattle(s.state.Galaxy, systemId)

				// Process health updates
				if len(healthUpdates) > 0 {
					s.tickMessage.HealthUpdates = append(s.tickMessage.HealthUpdates, healthUpdates...)
				}

				// Process destroyed fleets
				if len(fleetsDestroyed) > 0 {
					s.tickMessage.FleetDestroyed = append(s.tickMessage.FleetDestroyed, fleetsDestroyed...)
				}

				if battleActive {
					// Battle continues - keep the system in the battle list
					remainingBattlingSystems = append(remainingBattlingSystems, systemId)

					// Ensure the system remains marked as contested (owner="none") while battle is active
					system := s.state.Galaxy.Systems[systemId]
					if system.Owner != "none" {
						ownerChange, _ := s.state.SetSystemOwner(systemId, "none")
						if ownerChange != nil {
							s.tickMessage.SystemOwnerChanges = append(s.tickMessage.SystemOwnerChanges, ownerChange)
						}
					}
				} else {
					// Battle has ended - determine the new owner
					if newOwner != "" {
						ownerChange, err := s.state.SetSystemOwner(systemId, newOwner)
						if err == nil && ownerChange != nil {
							s.tickMessage.SystemOwnerChanges = append(s.tickMessage.SystemOwnerChanges, ownerChange)
						}
					}
				}
			}

			// Clean up: remove any duplicate system owner changes, keeping only the last one for each system
			if len(s.tickMessage.SystemOwnerChanges) > 1 {
				// Map to store the latest owner change per system
				latestChanges := make(map[int32]*pb.SystemOwnerChange)

				// Process changes in order, so the latest one will overwrite earlier ones
				for _, change := range s.tickMessage.SystemOwnerChanges {
					latestChanges[change.SystemId] = change
				}

				// Convert map back to slice
				updatedChanges := make([]*pb.SystemOwnerChange, 0, len(latestChanges))
				for _, change := range latestChanges {
					updatedChanges = append(updatedChanges, change)
				}

				// Replace the changes list with the deduplicated version
				s.tickMessage.SystemOwnerChanges = updatedChanges
			}

			s.state.battlingSystems = remainingBattlingSystems

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
			s.state.movedFleets = []int32{}

			s.state.mu.Unlock()
			s.mu.Unlock()
		}
	}
}
