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
	defer s.mu.Unlock()

	_, err := s.state.AddPlayer(req.Username)
	if err != nil {
		return nil, err
	}
	log.Printf("Player joined: %s", req.Username)

	return &pb.JoinResponse{
		PlayerCount: s.state.PlayerCount,
		Players:     s.state.Players,
		Started:     s.state.Started,
	}, nil
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
			if err := stream.Send(&pb.TickUpdate{
				PlayerCount: s.state.PlayerCount,
				Players:     s.state.Players,
				Started:     s.state.Started,
			}); err != nil {
				log.Printf("Error sending tick: %v", err)
				return err
			}
			log.Printf("Tick sent to client")
		}
	}
}
