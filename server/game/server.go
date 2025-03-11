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
}

func NewServer() *Server {
	s := &Server{
		ticker:  time.NewTicker(5 * time.Second),
		clients: make(map[string]pb.Game_SubscribeToTicksServer),
		state:   NewState(),
	}
	return s
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
	defer func() {
		s.mu.Lock()
		delete(s.clients, req.Username)
		s.mu.Unlock()
		s.state.RemovePlayer(req.Username)
		log.Printf("Client disconnected: %s", req.Username)
	}()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-s.ticker.C:
			log.Printf("Tick received from ticker")

			s.mu.Lock()
			update := &pb.TickUpdate{
				PlayerCount: s.state.PlayerCount,
				Players:     s.state.Players,
				Started:     s.state.Started,
			}
			s.mu.Unlock()

			if err := stream.Send(&pb.TickUpdate{
				PlayerCount: update.PlayerCount,
				Players:     update.Players,
				Started:     update.Started,
				Galaxy:      nil,
			}); err != nil {
				log.Printf("Error sending tick: %v", err)
				return err
			}
			log.Printf("Tick sent to client")
		}
	}
}
