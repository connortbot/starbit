package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	pb "starbit/proto"
	fleets "starbit/server/game/fleets"
	galaxy "starbit/server/game/galaxy"

	"github.com/quic-go/quic-go"
)

type UDPServer struct {
	listener  *quic.Listener
	state     *State
	tcpServer *Server // reference to TCP server for state syncing
	ticker    *time.Ticker
	clients   map[quic.Connection]map[quic.Stream]string // map conn->streams->username
	mu        sync.Mutex
	done      chan struct{}

	tickMessage *pb.TickMsg
}

func NewUDPServer(listener *quic.Listener) *UDPServer {
	return &UDPServer{
		listener:    listener,
		ticker:      time.NewTicker(500 * time.Millisecond),
		clients:     make(map[quic.Connection]map[quic.Stream]string),
		done:        make(chan struct{}),
		tickMessage: &pb.TickMsg{},
	}
}

// shared state setter
func (s *UDPServer) SetState(state *State) {
	s.state = state
}

func (s *UDPServer) SetTCPServer(tcpServer *Server) {
	s.tcpServer = tcpServer
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
				s.mu.Lock()
				ownerChange, err := s.state.MoveFleet(msg.Username, msg.TickMsg.FleetMovements[0])
				if err != nil {
					s.mu.Unlock()
					log.Printf("Error moving fleet: %v", err)

					errorMsg := ServerMessage{
						Type:     "error",
						Content:  err.Error(),
						Username: msg.Username,
					}

					jsonError, _ := json.Marshal(errorMsg)
					stream.Write(jsonError)
				} else {
					s.tickMessage.FleetMovements = append(s.tickMessage.FleetMovements, msg.TickMsg.FleetMovements[0])

					if ownerChange != nil {
						s.tickMessage.SystemOwnerChanges = append(s.tickMessage.SystemOwnerChanges, ownerChange)
					}

					s.mu.Unlock()
				}
			}
		case "fleet_creation":
			if msg.TickMsg != nil && len(msg.TickMsg.FleetCreations) > 0 {
				s.mu.Lock()

				systemId := msg.TickMsg.FleetCreations[0].SystemId
				if systemId < 0 || systemId >= int32(len(s.state.Galaxy.Systems)) {
					s.mu.Unlock()
					errorMsg := ServerMessage{
						Type:     "error",
						Content:  fmt.Sprintf("Invalid system ID: %d", systemId),
						Username: msg.Username,
					}
					jsonError, _ := json.Marshal(errorMsg)
					stream.Write(jsonError)
					continue
				}

				system := s.state.Galaxy.Systems[systemId]
				if system.Owner != msg.Username {
					s.mu.Unlock()
					errorMsg := ServerMessage{
						Type:     "error",
						Content:  fmt.Sprintf("You do not own system %d", systemId),
						Username: msg.Username,
					}
					jsonError, _ := json.Marshal(errorMsg)
					stream.Write(jsonError)
					continue
				}

				const fleetCost = 1000
				currentGES := s.state.GetPlayerGES(msg.Username)
				if currentGES < fleetCost {
					s.mu.Unlock()
					errorMsg := ServerMessage{
						Type:     "error",
						Content:  fmt.Sprintf("Not enough GES. Required: %d, Available: %d", fleetCost, currentGES),
						Username: msg.Username,
					}
					jsonError, _ := json.Marshal(errorMsg)
					stream.Write(jsonError)
					continue
				}

				newFleetId := s.state.nextFleetID
				s.state.nextFleetID++

				newFleet := fleets.NewFleet(newFleetId, msg.Username, galaxy.StartingFleetAttack, galaxy.StartingFleetHealth)
				galaxy.AddFleetToSystem(s.state.Galaxy, systemId, newFleet)

				newGES := s.state.AdjustPlayerGES(msg.Username, -fleetCost)

				fleetCreation := msg.TickMsg.FleetCreations[0]
				fleetCreation.FleetId = newFleetId
				fleetCreation.Owner = msg.Username
				fleetCreation.Attack = galaxy.StartingFleetAttack
				fleetCreation.Health = galaxy.StartingFleetHealth
				s.tickMessage.FleetCreations = append(s.tickMessage.FleetCreations, fleetCreation)

				s.tickMessage.GesUpdates = append(s.tickMessage.GesUpdates, &pb.GESUpdate{
					Owner:  msg.Username,
					Amount: newGES,
				})

				s.mu.Unlock()
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
			if !s.state.Started {
				continue
			}

			s.mu.Lock()
			if len(s.clients) == 0 {
				s.mu.Unlock()
				continue
			}
			s.state.mu.Lock()

			remainingBattlingSystems := []int32{}

			log.Printf("Processing %d battling systems: %v", len(s.state.battlingSystems), s.state.battlingSystems)

			systemCount := make(map[int32]int)
			for _, id := range s.state.battlingSystems {
				systemCount[id]++
			}

			for id, count := range systemCount {
				if count > 1 {
					log.Printf("WARNING: System %d appears %d times in battlingSystems!", id, count)
				}
			}

			processedSystemIds := make(map[int32]bool)

			for _, systemId := range s.state.battlingSystems {
				if processedSystemIds[systemId] {
					log.Printf("Skipping duplicate processing of system %d", systemId)
					continue
				}

				processedSystemIds[systemId] = true

				// Process battles
				log.Printf("Processing battle for system %d", systemId)
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

			// check for win conditions
			var victor string
			totalSystems := s.state.Galaxy.Width * s.state.Galaxy.Height
			for player, systems := range s.state.ownedSystems {
				if int32(len(systems)) == totalSystems {
					victor = player
					break
				}
			}

			s.state.battlingSystems = remainingBattlingSystems

			baseTickMsg := &pb.TickMsg{
				FleetMovements:     s.tickMessage.FleetMovements,
				HealthUpdates:      s.tickMessage.HealthUpdates,
				FleetDestroyed:     s.tickMessage.FleetDestroyed,
				SystemOwnerChanges: s.tickMessage.SystemOwnerChanges,
				FleetCreations:     s.tickMessage.FleetCreations,
			}

			// Handle victory case
			if victor != "" {
				victory := &pb.GameVictory{
					Winner: victor,
				}
				log.Printf("GAME OVER: %s has conquered the galaxy!", victor)

				baseTickMsg.Victory = victory
				s.state.mu.Unlock()
				newState := NewState()
				newState.Started = false
				s.state = newState

				if s.tcpServer != nil {
					s.tcpServer.SetState(newState)
					log.Printf("Game reset: Updated both UDP and TCP servers with new state (Started=%v)", newState.Started)
				} else {
					log.Printf("Warning: TCP server reference not available, only UDP state was reset")
				}

				// Send the final game state with victory notif to all clients
				for _, streams := range s.clients {
					for stream, username := range streams {
						update := ServerMessage{
							Type:    "tick",
							TickMsg: baseTickMsg,
						}

						jsonUpdate, _ := json.Marshal(update)
						if _, err := stream.Write(jsonUpdate); err != nil {
							log.Printf("Error sending victory notification to %s: %v", username, err)
						}
					}
				}

				s.tickMessage = &pb.TickMsg{}
				s.mu.Unlock()

				// start new tick with new state
				continue
			}

			for _, streams := range s.clients {
				for stream, username := range streams {
					clientTickMsg := &pb.TickMsg{
						FleetMovements:     baseTickMsg.FleetMovements,
						HealthUpdates:      baseTickMsg.HealthUpdates,
						FleetDestroyed:     baseTickMsg.FleetDestroyed,
						SystemOwnerChanges: baseTickMsg.SystemOwnerChanges,
						FleetCreations:     baseTickMsg.FleetCreations,
						Victory:            baseTickMsg.Victory,
					}

					if username != "" {
						newGES := s.state.AdjustPlayerGES(username, gesPerTick)
						clientTickMsg.GesUpdates = append(clientTickMsg.GesUpdates, &pb.GESUpdate{
							Owner:  username,
							Amount: newGES,
						})
					}

					update := ServerMessage{
						Type:    "tick",
						TickMsg: clientTickMsg,
					}

					jsonUpdate, _ := json.Marshal(update)
					if _, err := stream.Write(jsonUpdate); err != nil {
						log.Printf("Error sending tick to %s: %v", username, err)
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
