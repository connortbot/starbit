package main

import (
	"fmt"
	"log"
	"os"

	"starbit/client/game"
	ui "starbit/client/ui"
	pb "starbit/proto"

	tea "github.com/charmbracelet/bubbletea"
)

var debugLog *log.Logger

func init() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug"+os.Getenv("USER")+".log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		debugLog = log.New(f, "", log.LstdFlags)
	} else {
		debugLog = log.New(os.Stderr, "", log.LstdFlags)
	}
}

type ControlMode string

const (
	CommandMode ControlMode = "Command"
	InspectMode ControlMode = "Inspect"
	ExploreMode ControlMode = "Explore"
)

type model struct {
	username    string
	err         error
	client      *game.Client
	started     bool
	playerCount int32
	players     map[string]*pb.Player
	joined      bool
	galaxy      *pb.GalaxyState

	command   string
	inspector *ui.ScrollingViewport

	// selection coordinates for galaxy nav
	selectedX int32
	selectedY int32

	controlMode ControlMode
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.started {
		return m.HandleMenu(msg)
	} else {
		return m.HandleGame(msg)
	}
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress Ctrl+C to quit\n", m.err)
	}
	if !m.joined {
		return fmt.Sprintf("Enter your name: %s\n", m.username)
	}
	selectedSystemIndex := m.selectedY*m.galaxy.Width + m.selectedX
	return ui.RenderGameScreen(
		m.username,
		m.players,
		m.started,
		m.galaxy,
		m.command,
		m.inspector,
		m.selectedX,
		m.selectedY,
		m.galaxy.Systems[selectedSystemIndex],
		string(m.controlMode),
	)
}

func listenForTicks(client *game.Client, p *tea.Program) {
	for {
		if client.Stream == nil {
			continue
		}
		tick, err := client.Stream.Recv()
		if err != nil {
			debugLog.Printf("Error receiving tick: %v", err)
			return
		}
		tickMsg := game.TickMsg{
			PlayerCount: tick.PlayerCount,
			Players:     tick.Players,
			Started:     tick.Started,
			Galaxy:      tick.Galaxy,
		}
		debugLog.Printf("Tick: %v", tickMsg)
		p.Send(tickMsg)
	}
}

func main() {
	client := game.NewClient()
	if err := client.Connect(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	p := tea.NewProgram(model{
		client: client,
	}, tea.WithAltScreen())
	go listenForTicks(client, p)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
