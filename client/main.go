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
		f, err := tea.LogToFile("debug" + os.Getenv("USER") + ".log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		debugLog = log.New(f, "", log.LstdFlags)
	} else {
		debugLog = log.New(os.Stderr, "", log.LstdFlags)
	}
}

type model struct {
	username    string
	err         error
	client      *game.Client
	started     bool
	playerCount int32
	players     map[string]*pb.Player
	tickChan    chan game.TickMsg
	joined      bool
	galaxy      *pb.GalaxyState

	command string
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
	return ui.RenderGameScreen(m.username, m.players, m.started, m.galaxy, m.command)
}

func handleTicks(client *game.Client, tickChan chan<- game.TickMsg) {
	defer close(tickChan)
	for {
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
		tickChan <- tickMsg
	}
}

func waitForTicks(tickChan <-chan game.TickMsg) tea.Cmd {
	return func() tea.Msg {
		return <-tickChan
	}
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
