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

type model struct {
	username    string
	err         error
	client      *game.Client
	started     bool
	playerCount int32
	players     map[string]*pb.Player
	tickChan    chan game.TickMsg
	joined      bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.client != nil {
				m.client.Close()
			}
			return m, tea.Quit
		case "enter":
			if m.username != "" && !m.started {
				client := game.NewClient()
				if err := client.Connect(); err != nil {
					m.err = err
					return m, nil
				}
				if err := client.SubscribeToTicks(m.username); err != nil {
					client.Close()
					m.err = err
					return m, nil
				}
				m.client = client
				m.tickChan = make(chan game.TickMsg)
				go handleTicks(client, m.tickChan)
				resp, err := client.JoinGame(m.username)
				if err != nil {
					client.Close()
					m.err = err
					return m, nil
				}
				m.joined = true
				m.playerCount = resp.PlayerCount
				m.players = resp.Players
				m.started = resp.Started
				return m, waitForTicks(m.tickChan)
			}
		case "backspace":
			if len(m.username) > 0 && !m.started {
				m.username = m.username[:len(m.username)-1]
			}
		default:
			if len(msg.String()) == 1 && !m.started {
				m.username += msg.String()
			}
		}
	case game.TickMsg:
		m.playerCount = msg.PlayerCount
		m.started = msg.Started
		m.players = msg.Players
		return m, waitForTicks(m.tickChan)
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress Ctrl+C to quit\n", m.err)
	}
	if !m.joined {
		return fmt.Sprintf("Enter your name: %s\n", m.username)
	}
	return ui.RenderPlayerList(m.username, m.players)
}

func handleTicks(client *game.Client, tickChan chan<- game.TickMsg) {
	defer close(tickChan)
	for {
		tick, err := client.Stream.Recv()
		if err != nil {
			debugLog.Printf("Error receiving tick: %v", err)
			return
		}
		debugLog.Printf("Received tick from server: %v", tick)
		tickMsg := game.TickMsg{
			PlayerCount: tick.PlayerCount,
			Players:     tick.Players,
			Started:     tick.Started,
		}
		debugLog.Printf("Sending tick to channel: %v", tickMsg)
		tickChan <- tickMsg
	}
}

func waitForTicks(tickChan <-chan game.TickMsg) tea.Cmd {
	return func() tea.Msg {
		return <-tickChan
	}
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
