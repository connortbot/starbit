package game

import (
	"context"
	"log"
	"fmt"

	pb "starbit/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GameMsg struct {
	PlayerCount int32
	Players     map[string]*pb.Player
	Started     bool
	Galaxy      *pb.GalaxyState
}

type Client struct {
	username string
	ip       string
	tcpPort  string
	conn     *grpc.ClientConn
	client   pb.GameClient
	Stream   pb.Game_MaintainConnectionClient
	gameCh   chan GameMsg // channel for game updates from TCP
}

func NewClient() *Client {
	return &Client{
		gameCh: make(chan GameMsg, 10),
	}
}

func (c *Client) SetConnectionInfo(ip string, tcpPort string) {
	c.ip = ip
	c.tcpPort = tcpPort
}

func (c *Client) Connect() error {
	if c.ip == "" || c.tcpPort == "" {
		return fmt.Errorf("ip and tcpPort must be set")
	}
	// always use conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(c.ip+":"+c.tcpPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

// receives messages from the TCP server and forwards them to the gameCh
func (c *Client) receiveUpdates() {
	for {
		update, err := c.Stream.Recv()
		if err != nil {
			log.Printf("Error receiving from TCP stream: %v", err)
			return
		}

		log.Printf("Received TCP update: started=%v, galaxy=%v", update.Started, update.Galaxy != nil)

		// convert to GameMsg and send to channel
		c.gameCh <- GameMsg{
			PlayerCount: update.PlayerCount,
			Players:     update.Players,
			Started:     update.Started,
			Galaxy:      update.Galaxy,
		}
	}
}

func (c *Client) GetUpdateChannel() <-chan GameMsg {
	return c.gameCh
}

func (c *Client) Close() {
	if c.Stream != nil {
		c.Stream.CloseSend()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
