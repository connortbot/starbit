package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "starbit/proto"

	"google.golang.org/grpc"
)

const (
	galaxyWidth  = 10
	galaxyHeight = 10
	maxPlayers   = 2
)

type gameState struct {
	playerCount int32
	players     []*pb.Player
	galaxy      *pb.GalaxyState
	gameStarted bool
	mu          sync.Mutex
}

func newGameState() *gameState {
	systems := make([]*pb.System, 0, galaxyWidth*galaxyHeight)
	for y := 0; y < galaxyHeight; y++ {
		for x := 0; x < galaxyWidth; x++ {
			systems = append(systems, &pb.System{
				X: int32(x),
				Y: int32(y),
			})
		}
	}

	return &gameState{
		playerCount: 0,
		players:     make([]*pb.Player, 0),
		galaxy: &pb.GalaxyState{
			Systems: systems,
		},
		gameStarted: false,
	}
}

func (g *gameState) addPlayer(name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.players) >= maxPlayers {
		return fmt.Errorf("game is full (max %d players)", maxPlayers)
	}

	player := &pb.Player{
		Id:   fmt.Sprintf("player_%d", len(g.players)+1),
		Name: name,
	}
	g.players = append(g.players, player)
	g.playerCount = int32(len(g.players))
	log.Printf("Player joined: %s (ID: %s)", name, player.Id)

	// Start game when we hit max players
	if len(g.players) == maxPlayers {
		g.gameStarted = true
		log.Printf("Game started with %d players!", maxPlayers)
	}

	return nil
}

func (g *gameState) getState() *pb.GameState {
	g.mu.Lock()
	defer g.mu.Unlock()

	return &pb.GameState{
		PlayerCount: g.playerCount,
		Players:     g.players,
		Galaxy:      g.galaxy,
		GameStarted: g.gameStarted,
	}
}

type server struct {
	pb.UnimplementedGameServer
	gameState *gameState
}

func newServer() *server {
	return &server{
		gameState: newGameState(),
	}
}

func (s *server) JoinGame(ctx context.Context, req *pb.JoinRequest) (*pb.Empty, error) {
	if err := s.gameState.addPlayer(req.Username); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *server) SubscribeToTicks(empty *pb.Empty, stream pb.Game_SubscribeToTicksServer) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			state := s.gameState.getState()
			if !state.GameStarted {
				continue
			}
			if err := stream.Send(&pb.TickUpdate{State: state}); err != nil {
				log.Printf("Error sending tick: %v", err)
				return err
			}
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGameServer(s, newServer())
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
