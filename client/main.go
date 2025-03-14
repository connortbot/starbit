package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"starbit/client/game"
	ui "starbit/client/ui"
	pb "starbit/proto"

	tea "github.com/charmbracelet/bubbletea"
)

var program *tea.Program

type ControlMode string

const (
	CommandMode   ControlMode = "Command"
	InspectMode   ControlMode = "Inspect"
	ExploreMode   ControlMode = "Explore"
	FleetListMode ControlMode = "FleetList"
)

type InputMode string

const (
	IPMode       InputMode = "IP"
	UsernameMode InputMode = "Username"
)

type model struct {
	firstScreen bool

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
	fleetList *ui.ScrollingViewport

	gameLogger *ui.GameLogger
	logWindow  *ui.ScrollingViewport

	tickCount int32

	// selection coordinates for galaxy nav
	selectedX int32
	selectedY int32

	controlMode ControlMode
	inputMode   InputMode
	ipAddress   string
	connected   bool

	udpClient *game.UDPClient

	ownedFleets   []*pb.Fleet
	fleetLocations map[int32]int32
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var result model
	var cmd tea.Cmd

	if m.firstScreen {
		result, cmd = m.HandleFirstScreen(msg)
		return result, cmd
	}

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

	if m.firstScreen {
		return ui.RenderFirstScreen()
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
		m.fleetList,
		m.selectedX,
		m.selectedY,
		m.galaxy.Systems[selectedSystemIndex],
		string(m.controlMode),
		m.gesAmount,
		m.tickCount,
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
		log.Printf("TCP Update: players=%d, started=%v, hasGalaxy=%v",
			len(update.Players), update.Started, update.Galaxy != nil)
		p.Send(update)
	}
}

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug"+os.Getenv("USER")+".log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		log.SetOutput(io.Discard)
	}
	log.Println("Initializing game...")
	log.Println("Initializing UDP client...")
	udpClient := game.NewUDPClient()
	log.Println("initializing TCP client...")
	client := game.NewClient()

	gameLogger := ui.NewGameLogger(100) // store up to 100 log messages
	gameLogger.AddSystemMessage("Welcome to Starbit! Enter your username to join.")
	logWindow := ui.NewLogWindow(gameLogger, ui.LogBoxWidth, ui.LogBoxHeight)

	program = tea.NewProgram(model{
		firstScreen: true,
		connected:   false,
		client:      client,
		udpClient:   udpClient,
		gameLogger:  gameLogger,
		logWindow:   logWindow,
		inputMode:   UsernameMode,

		ipAddress: "localhost",
	}, tea.WithAltScreen())

	// run the UI
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}
}
