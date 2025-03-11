package game

import (
	"context"
	"log"
	"sync"
	"time"

	pb "starbit/proto"
)

type Server struct {
	pb.UnimplementedGameServer
	ticker  *time.Ticker
	mu      sync.Mutex
	clients map[string]pb.Game_SubscribeToTicksServer
	state   *State
	done    chan struct{}
}

func NewServer() *Server {
	s := &Server{
		ticker:  time.NewTicker(5 * time.Second),
		clients: make(map[string]pb.Game_SubscribeToTicksServer),
		state:   NewState(),
		done:    make(chan struct{}),
	}
	go s.broadcastTicks()
	return s
}

func (s *Server) broadcastTicks() {
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			s.mu.Lock()
			update := &pb.TickUpdate{
				PlayerCount: s.state.PlayerCount,
				Players:     s.state.Players,
				Started:     s.state.Started,
				Galaxy:      nil,
			}
			var clientsToRemove []string

			for username, client := range s.clients {
				if err := client.Send(update); err != nil {
					log.Printf("Error sending tick to client %s: %v", username, err)
					clientsToRemove = append(clientsToRemove, username)
				}
			}

			for _, username := range clientsToRemove {
				delete(s.clients, username)
				s.state.RemovePlayer(username)
				log.Printf("Removed disconnected client: %s", username)
			}
			s.mu.Unlock()
		}
	}
}

func (s *Server) JoinGame(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	s.mu.Lock()
	full, err := s.state.AddPlayer(req.Username)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	if full {
		s.state.InitializeGame(s.state.Players)
	}

	response := &pb.JoinResponse{
		PlayerCount: s.state.PlayerCount,
		Players:     s.state.Players,
		Started:     s.state.Started,
		Galaxy:      s.state.Galaxy,
	}

	s.mu.Unlock()

	if full {
		log.Printf("Amount of clients: %d", len(s.clients))
		for _, client := range s.clients {
			client.Send(&pb.TickUpdate{
				PlayerCount: response.PlayerCount,
				Players:     response.Players,
				Started:     response.Started,
				Galaxy:      response.Galaxy,
			})
		}
	}
	log.Printf("Player joined: %s", req.Username)

	return response, nil
}

func (s *Server) SubscribeToTicks(req *pb.SubscribeRequest, stream pb.Game_SubscribeToTicksServer) error {
	s.mu.Lock()
	s.clients[req.Username] = stream
	s.mu.Unlock()

	// wait for client disconnection
	<-stream.Context().Done()

	s.mu.Lock()
	delete(s.clients, req.Username)
	s.state.RemovePlayer(req.Username)
	s.mu.Unlock()
	log.Printf("Client disconnected: %s", req.Username)

	return nil
}
