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
var program *tea.Program

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

type InputMode string

const (
	IPMode       InputMode = "IP"
	UsernameMode InputMode = "Username"
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

	gesAmount int32

	command   string
	inspector *ui.ScrollingViewport

	gameLogger *ui.GameLogger
	logWindow  *ui.ScrollingViewport

	// selection coordinates for galaxy nav
	selectedX int32
	selectedY int32

	controlMode ControlMode
	inputMode   InputMode
	ipAddress   string
	connected   bool

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
		return ui.RenderJoinScreen(m.username, m.logWindow, m.ipAddress, m.connected, string(m.inputMode))
	}

	selectedSystemIndex := m.selectedY*m.galaxy.Width + m.selectedX
	return ui.RenderGameScreen(
		m.username,
		m.players,
		m.started,
		m.galaxy,
		m.command,
		m.inspector,
		m.logWindow,
		m.selectedX,
		m.selectedY,
		m.galaxy.Systems[selectedSystemIndex],
		string(m.controlMode),
		m.gesAmount,
	)
}

// receives game state updates from the UDP server
func listenForUDPTicks(udpClient *game.UDPClient, p *tea.Program) {
	tickCh := udpClient.GetTickChannel()
	for tick := range tickCh {
		p.Send(tick)
	}
}

func listenForUDPErrors(udpClient *game.UDPClient, p *tea.Program) {
	errorCh := udpClient.GetErrorChannel()
	for errorMsg := range errorCh {
		p.Send(errorMsg)
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
	debugLog.Println("initializing TCP client...")
	client := game.NewClient()

	gameLogger := ui.NewGameLogger(100) // store up to 100 log messages
	gameLogger.AddSystemMessage("Welcome to Starbit! Enter your username to join.")
	logWindow := ui.NewLogWindow(gameLogger, ui.LogBoxWidth, ui.LogBoxHeight)

	program = tea.NewProgram(model{
		connected:  false,
		client:     client,
		udpClient:  udpClient,
		gameLogger: gameLogger,
		logWindow:  logWindow,
		inputMode:  UsernameMode,

		ipAddress: "3.133.113.171",
	}, tea.WithAltScreen())

	// run the UI
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}
}
