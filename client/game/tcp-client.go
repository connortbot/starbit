package game

import (
	"context"
	"log"

	pb "starbit/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TickMsg struct {
	PlayerCount int32
	Players     map[string]*pb.Player
	Started     bool
	Galaxy      *pb.GalaxyState
}

type Client struct {
	username string
	conn     *grpc.ClientConn
	client   pb.GameClient
	Stream   pb.Game_MaintainConnectionClient
	tickCh   chan TickMsg // channel for game updates from TCP
}

func NewClient() *Client {
	return &Client{
		tickCh: make(chan TickMsg, 10),
	}
}

func (c *Client) Connect() error {
	// always use conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	c.client = pb.NewGameClient(conn)
	c.conn = conn
	return nil
}

func (c *Client) JoinGame(name string) (*pb.JoinResponse, error) {
	c.username = name
	resp, err := c.client.JoinGame(context.Background(), &pb.JoinRequest{
		Username: name,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) MaintainConnection(name string) error {
	c.username = name
	stream, err := c.client.MaintainConnection(context.Background(), &pb.ConnectionRequest{
		Username: name,
	})
	if err != nil {
		return err
	}

	c.Stream = stream

	// goroutine to receive messages from the server
	go c.receiveUpdates()

	return nil
}

// receives messages from the TCP server and forwards them to the tickCh
func (c *Client) receiveUpdates() {
	for {
		update, err := c.Stream.Recv()
		if err != nil {
			log.Printf("Error receiving from TCP stream: %v", err)
			return
		}

		log.Printf("Received TCP update: started=%v, galaxy=%v", update.Started, update.Galaxy != nil)

		// convert to TickMsg and send to channel
		c.tickCh <- TickMsg{
			PlayerCount: update.PlayerCount,
			Players:     update.Players,
			Started:     update.Started,
			Galaxy:      update.Galaxy,
		}
	}
}

func (c *Client) GetUpdateChannel() <-chan TickMsg {
	return c.tickCh
}

func (c *Client) Close() {
	if c.Stream != nil {
		c.Stream.CloseSend()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
