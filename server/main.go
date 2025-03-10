package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	pb "starbit/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGameServer
	messageCount int32
	mu           sync.Mutex
}

// SendMessage implements the Game service
func (s *server) SendMessage(ctx context.Context, msg *pb.GameMessage) (*pb.Empty, error) {
	s.mu.Lock()
	s.messageCount++
	s.mu.Unlock()
	log.Printf("Received message: %s (Total: %d)", msg.Content, s.messageCount)
	return &pb.Empty{}, nil
}

// SubscribeToTicks implements the Game service
func (s *server) SubscribeToTicks(empty *pb.Empty, stream pb.Game_SubscribeToTicksServer) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			s.mu.Lock()
			count := s.messageCount
			s.messageCount = 0
			s.mu.Unlock()

			if err := stream.Send(&pb.TickUpdate{MessageCount: count}); err != nil {
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
	pb.RegisterGameServer(s, &server{})
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
