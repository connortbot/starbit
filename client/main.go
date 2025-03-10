package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	pb "starbit/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create gRPC client
	gameClient := pb.NewGameClient(conn)

	// Subscribe to ticks
	stream, err := gameClient.SubscribeToTicks(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("could not subscribe to ticks: %v", err)
	}

	// Start a goroutine to receive ticks
	go func() {
		for {
			tick, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Error receiving tick: %v", err)
				return
			}
			log.Printf("Tick received! Messages in last tick: %d", tick.MessageCount)
		}
	}()

	// Send messages every second
	for {
		_, err := gameClient.SendMessage(context.Background(), &pb.GameMessage{
			Content: "Hello from " + *name,
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		time.Sleep(time.Second)
	}
}
