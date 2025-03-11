package game

import (
	"context"

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
	Stream   pb.Game_SubscribeToTicksClient
}

func NewClient() *Client {
	return &Client{}
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

func (c *Client) SubscribeToTicks(name string) error {
	c.username = name
	stream, err := c.client.SubscribeToTicks(context.Background(), &pb.SubscribeRequest{
		Username: name,
	})
	if err != nil {
		return err
	}
	c.Stream = stream
	return nil
}

func (c *Client) Close() {
	if c.Stream != nil {
		c.Stream.CloseSend()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
