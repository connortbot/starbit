package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	pb "starbit/proto"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var debugLog *log.Logger

func init() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		debugLog = log.New(f, "", log.LstdFlags)
	} else {
		debugLog = log.New(os.Stderr, "", log.LstdFlags)
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

type model struct {
	grpcConn   *grpc.ClientConn
	gameClient pb.GameClient
	stream     pb.Game_SubscribeToTicksClient
	name       string
	nameInput  string
	state      string
	err        error
}

func initialModel() model {
	return model{
		state: "asking_name",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.state == "asking_name" && m.nameInput != "" {
				m.name = m.nameInput
				m.state = "connecting"
				return m, connect()
			}
		case "ctrl+c":
			return m, tea.Quit
		default:
			if m.state == "asking_name" && len(msg.String()) == 1 {
				m.nameInput += msg.String()
			}
		}
	case errorMsg:
		m.err = msg
		m.state = "error"
	case connectedMsg:
		m.grpcConn = msg.conn
		m.gameClient = msg.client
		m.stream = msg.stream
		m.state = "joining"
		return m, m.joinGame()
	case joinedMsg:
		m.state = "connected"
		go m.handleTicks()
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case "asking_name":
		return fmt.Sprintf("Enter your name: %s\n", m.nameInput)
	case "connecting":
		return "Connecting to server...\n"
	case "joining":
		return "Joining game...\n"
	case "error":
		return fmt.Sprintf("Error: %v\nPress Ctrl+C to quit", m.err)
	case "connected":
		return fmt.Sprintf("Connected as: %s\nWaiting for players...\nPress Ctrl+C to quit", m.name)
	default:
		return "Unknown state\n"
	}
}

type errorMsg struct{ err error }

func (e errorMsg) Error() string {
	return e.err.Error()
}

type connectedMsg struct {
	conn   *grpc.ClientConn
	client pb.GameClient
	stream pb.Game_SubscribeToTicksClient
}

type joinedMsg struct{}

func connect() tea.Cmd {
	return func() tea.Msg {
		conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return errorMsg{err}
		}

		gameClient := pb.NewGameClient(conn)
		stream, err := gameClient.SubscribeToTicks(context.Background(), &pb.Empty{})
		if err != nil {
			return errorMsg{err}
		}

		return connectedMsg{
			conn:   conn,
			client: gameClient,
			stream: stream,
		}
	}
}

func (m model) joinGame() tea.Cmd {
	return func() tea.Msg {
		_, err := m.gameClient.JoinGame(context.Background(), &pb.JoinRequest{
			Username: m.name,
		})
		if err != nil {
			return errorMsg{err}
		}
		return joinedMsg{}
	}
}

func (m model) handleTicks() {
	for {
		tick, err := m.stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			debugLog.Printf("Error receiving tick: %v", err)
			return
		}
		if tick.State.GameStarted {
			debugLog.Printf("Game started! Players: %d", tick.State.PlayerCount)
			for _, p := range tick.State.Players {
				debugLog.Printf("Player: %s (ID: %s)", p.Name, p.Id)
			}
		} else {
			debugLog.Printf("Waiting for players: %d", tick.State.PlayerCount)
		}
	}
}
