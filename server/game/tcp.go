package game

import (
	"context"
	"log"
	"sync"

	"errors"
	pb "starbit/proto"
)

type Server struct {
	pb.UnimplementedGameServer
	mu      sync.Mutex
	clients map[string]pb.Game_MaintainConnectionServer
	state   *State
	done    chan struct{}
}

func NewServer() *Server {
	return &Server{
		clients: make(map[string]pb.Game_MaintainConnectionServer),
		done:    make(chan struct{}),
	}
}

func (s *Server) SetState(state *State) {
	s.state = state
}

func (s *Server) JoinGame(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	s.mu.Lock()

	if s.state.Started {
		s.mu.Unlock()
		return nil, errors.New("game already started")
	}

	full, err := s.state.AddPlayer(req.Username)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	// if game is now full, initialize it
	gameJustStarted := false
	if full {
		s.state.InitializeGame(s.state.Players)
		gameJustStarted = true
	}

	// prepare response with full galaxy data
	response := &pb.JoinResponse{
		PlayerCount: s.state.PlayerCount,
		Players:     s.state.Players,
		Started:     s.state.Started,
		Galaxy:      s.state.Galaxy,
	}

	s.mu.Unlock()
	log.Printf("Player joined: %s (started: %v)", req.Username, s.state.Started)

	// if game just started, notify all connected clients with galaxy data
	if gameJustStarted {
		log.Printf("Game just started, broadcasting to %d clients", len(s.clients))
		s.broadcastGameStart()
	}

	return response, nil
}

// broadcastGameStart sends the initial game state with galaxy data to all connected clients
func (s *Server) broadcastGameStart() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Broadcasting game start to %d clients", len(s.clients))
	for username, client := range s.clients {
		err := client.Send(&pb.GameUpdate{
			PlayerCount: s.state.PlayerCount,
			Players:     s.state.Players,
			Started:     s.state.Started,
			Galaxy:      s.state.Galaxy, // with galaxy data
		})
		if err != nil {
			log.Printf("Error sending initial state to %s: %v", username, err)
		} else {
			log.Printf("Sent initial game state to %s", username)
		}
	}
}

func (s *Server) MaintainConnection(req *pb.ConnectionRequest, stream pb.Game_MaintainConnectionServer) error {
	s.mu.Lock()
	s.clients[req.Username] = stream
	s.mu.Unlock()

	// wait for client disconnection
	<-stream.Context().Done()

	// handle client disconnection
	s.mu.Lock()
	delete(s.clients, req.Username)
	s.state.RemovePlayer(req.Username)
	s.mu.Unlock()
	log.Printf("Client disconnected: %s", req.Username)

	return nil
}
