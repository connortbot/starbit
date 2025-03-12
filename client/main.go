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

	udpClient *game.UDPClient
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var result model
	var cmd tea.Cmd

	if !m.started {
		result, cmd = m.HandleMenu(msg)
	} else {
		result, cmd = m.HandleGame(msg)
	}

	return result, cmd
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

// receives game state updates from the UDP server
func listenForUDPTicks(udpClient *game.UDPClient, p *tea.Program) {
	tickCh := udpClient.GetTickChannel()
	for tick := range tickCh {
		debugLog.Printf("UDP Tick: %v", tick)
		p.Send(tick)
	}
}

// receives game state updates from the TCP server
func listenForTCPUpdates(client *game.Client, p *tea.Program) {
	updateCh := client.GetUpdateChannel()
	for update := range updateCh {
		debugLog.Printf("TCP Update: players=%d, started=%v, hasGalaxy=%v",
			len(update.Players), update.Started, update.Galaxy != nil)
		p.Send(update)
	}
}

func main() {
	debugLog.Println("Initializing UDP client...")
	udpClient := game.NewUDPClient()
	if err := udpClient.Connect(); err != nil {
		fmt.Printf("Error connecting UDP client: %v\n", err)
		debugLog.Printf("UDP connection failed: %v", err)
		os.Exit(1)
	}
	debugLog.Println("UDP client connected successfully")

	// initialize TCP client for game joining and initial galaxy state
	debugLog.Println("initializing TCP client...")
	client := game.NewClient()
	if err := client.Connect(); err != nil {
		fmt.Printf("Error connecting TCP client: %v\n", err)
		os.Exit(1)
	}
	debugLog.Println("TCP client connected successfully")

	p := tea.NewProgram(model{
		client:    client,
		udpClient: udpClient,
	}, tea.WithAltScreen())

	// start listening for both UDP and TCP game updates
	go listenForUDPTicks(udpClient, p)
	go listenForTCPUpdates(client, p)

	// run the UI
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}
}
