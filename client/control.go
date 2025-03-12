package main

import (
	"starbit/client/game"

	"starbit/client/ui"

	"time"

	pb "starbit/proto"

	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) HandleMenu(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.client != nil {
				m.client.Close()
			}
			return m, tea.Quit
		case "T", "shift+t":
			m.inputMode = TCPPortMode
		case "U", "shift+u":
			m.inputMode = UDPPortMode
		case "I", "shift+i":
			m.inputMode = IPMode
		case "N", "shift+n":
			m.inputMode = UsernameMode
		case "enter":
			if !m.connected {
				readyToConnect := m.ipAddress != "" && m.tcpPort != "" && m.udpPort != "" && !m.started
				if !readyToConnect {
					m.gameLogger.AddSystemMessage("Please enter a IP address, TCP port, and UDP port to connect to the server.")
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					return m, nil
				}
				m.client.SetConnectionInfo(m.ipAddress, m.tcpPort)
				m.udpClient.SetConnectionInfo(m.ipAddress, m.udpPort)

				m.client.Connect()
				if err := m.client.Connect(); err != nil {
					debugLog.Printf("Error connecting TCP client: %v", err)
					m.gameLogger.AddSystemMessage(fmt.Sprintf("Error connecting to TCP client [%v] Double check your IP and TCP port!", err))
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					return m, nil
				}
				debugLog.Printf("TCP client connected successfully")

				if err := m.udpClient.Connect(); err != nil {
					debugLog.Printf("Error connecting UDP client: %v", err)
					m.gameLogger.AddSystemMessage(fmt.Sprintf("Error connecting to UDP client [%v] Double check your IP and UDP port!", err))
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					return m, nil
				}
				debugLog.Printf("UDP client connected successfully")

				go listenForUDPTicks(m.udpClient, program)
				go listenForUDPErrors(m.udpClient, program)
				go listenForTCPUpdates(m.client, program)
				debugLog.Printf("Listening for UDP ticks, errors, and TCP updates")

				m.connected = true
				return m, nil
			} else {
				debugLog.Printf("Joining Game")
				readyToJoin := m.username != ""
				if !readyToJoin {
					m.gameLogger.AddSystemMessage("Please enter a username to join the game.")
					return m, nil
				}
				if err := m.udpClient.Register(m.username); err != nil {
					m.err = err
					return m, nil
				}

				if err := m.client.MaintainConnection(m.username); err != nil {
					m.client.Close()
					m.err = err
					return m, nil
				}

				// short delay to ensure connection is established
				time.Sleep(100 * time.Millisecond)

				// then join the game to get the initial state
				resp, err := m.client.JoinGame(m.username)
				if err != nil {
					m.client.Close()
					m.err = err
					return m, nil
				}

				m.joined = true
				m.playerCount = resp.PlayerCount
				m.players = resp.Players
				m.started = resp.Started
				m.galaxy = resp.Galaxy
				m.inspector = ui.NewInspectWindow(ui.InspectorWidth, m.galaxy.Systems[0])
				m.controlMode = CommandMode

				m.gameLogger.AddSystemMessage(fmt.Sprintf("%s joined the game", m.username))
				m.logWindow = ui.NewLogWindow(m.gameLogger, ui.LogBoxWidth, ui.LogBoxHeight)

				if resp.Started {
					m.gameLogger.AddSystemMessage("Game started")
				}
				// log
				debugLog.Printf("Joined game successfully. Started: %v, Players: %d",
					m.started, m.playerCount)

				return m, nil
			}
		case "backspace":
			if m.inputMode == UsernameMode {
				if len(m.username) > 0 && !m.started {
					m.username = m.username[:len(m.username)-1]
				}
			} else if m.inputMode == IPMode {
				if len(m.ipAddress) > 0 && !m.started {
					m.ipAddress = m.ipAddress[:len(m.ipAddress)-1]
				}
			} else if m.inputMode == TCPPortMode {
				if len(m.tcpPort) > 0 && !m.started {
					m.tcpPort = m.tcpPort[:len(m.tcpPort)-1]
				}
			} else if m.inputMode == UDPPortMode {
				if len(m.udpPort) > 0 && !m.started {
					m.udpPort = m.udpPort[:len(m.udpPort)-1]
				}
			}
		default:
			if len(msg.String()) == 1 && !m.started {
				if m.inputMode == UsernameMode {
					m.username += msg.String()
				} else if m.inputMode == IPMode {
					m.ipAddress += msg.String()
				} else if m.inputMode == TCPPortMode {
					m.tcpPort += msg.String()
				} else if m.inputMode == UDPPortMode {
					m.udpPort += msg.String()
				}
			}
		}
	case game.GameMsg:
		m.playerCount = msg.PlayerCount
		m.started = msg.Started
		m.players = msg.Players
		if msg.Galaxy != nil {
			m.galaxy = msg.Galaxy
		}
		if msg.Started {
			m.gameLogger.AddSystemMessage("Game started")
		}
		return m, nil
	}
	return m, nil
}

func (m model) HandleGame(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.client != nil {
				m.client.Close()
			}
			return m, tea.Quit
		case "C", "shift+c":
			m.controlMode = CommandMode
		case "I", "shift+i":
			m.controlMode = InspectMode
		case "E", "shift+e":
			m.controlMode = ExploreMode
		case "up":
			if m.controlMode == InspectMode {
				m.inspector.ScrollUp()
			} else if m.controlMode == ExploreMode {
				if m.selectedY > 0 {
					m.selectedY--
				}
			}
		case "down":
			if m.controlMode == InspectMode {
				m.inspector.ScrollDown()
			} else if m.controlMode == ExploreMode {
				if m.galaxy != nil && m.selectedY < m.galaxy.Height-1 {
					m.selectedY++
				}
			}
		case "shift+up":
			if m.controlMode == InspectMode {
				m.inspector.ScrollToTop()
			}
		case "shift+down":
			if m.controlMode == InspectMode {
				m.inspector.ScrollToBottom()
			}
		case "left":
			if m.controlMode == ExploreMode {
				if m.selectedX > 0 {
					m.selectedX--
				}
			}
		case "right":
			if m.controlMode == ExploreMode {
				if m.galaxy != nil && m.selectedX < m.galaxy.Width-1 {
					m.selectedX++
				}
			}
		case "enter":
			if m.controlMode == CommandMode && m.command != "" {
				debugLog.Printf("Processing command: %s", m.command)
				result := game.ParseCommand(m.udpClient, m.command)
				if result.Success {
					debugLog.Printf("Command successful: %s", result.Message)
					m.gameLogger.AddCommand(m.username, m.command, true)
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					m.command = ""
				} else {
					debugLog.Printf("Command failed: %s", result.Message)
					m.gameLogger.AddCommand(m.username, m.command, false)
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
				}
			}
		case "backspace":
			if m.controlMode == CommandMode && len(m.command) > 0 {
				m.command = m.command[:len(m.command)-1]
			}
		default:
			if m.controlMode == CommandMode && len(msg.String()) == 1 {
				if strings.HasPrefix(m.command, "ERROR: ") {
					m.command = msg.String()
				} else {
					m.command += msg.String()
				}
			}
		}
	case game.GameMsg:
		m.playerCount = msg.PlayerCount
		m.started = msg.Started
		m.players = msg.Players
		if msg.Galaxy != nil {
			m.galaxy = msg.Galaxy
		}
		return m, nil
	case *pb.TickMsg:
		debugLog.Printf("UDP Tick: %s", string(msg.Message))
		if len(msg.FleetMovements) > 0 {
			debugLog.Printf("Received %d fleet movements", len(msg.FleetMovements))
			for _, movement := range msg.FleetMovements {
				err := game.MoveFleet(m.galaxy, movement)
				if err != nil {
					debugLog.Printf("Error processing fleet movement: %v", err)
				} else {
					debugLog.Printf("Fleet %d moved from system %d to system %d",
						movement.FleetId, movement.FromSystemId, movement.ToSystemId)
					m.gameLogger.AddFleetMovement(movement)
				}
			}
		}

		if len(msg.HealthUpdates) > 0 {
			debugLog.Printf("Received %d health updates", len(msg.HealthUpdates))
			for _, update := range msg.HealthUpdates {
				err := game.ApplyHealthUpdate(m.galaxy, update)
				if err != nil {
					debugLog.Printf("Error applying health update: %v", err)
				} else {
					debugLog.Printf("Fleet %d health updated to %d in system %d",
						update.FleetId, update.Health, update.SystemId)
					m.gameLogger.AddHealthUpdate(update)
				}
			}
		}

		if len(msg.FleetDestroyed) > 0 {
			debugLog.Printf("Received %d destroyed fleets", len(msg.FleetDestroyed))
			for _, destroyed := range msg.FleetDestroyed {
				err := game.ProcessFleetDestroyed(m.galaxy, destroyed)
				if err != nil {
					debugLog.Printf("Error processing destroyed fleet: %v", err)
				} else {
					debugLog.Printf("Fleet %d was destroyed in system %d",
						destroyed.FleetId, destroyed.SystemId)
					m.gameLogger.AddFleetDestroyed(destroyed)
				}
			}
		}

		if len(msg.SystemOwnerChanges) > 0 {
			debugLog.Printf("Received %d system owner changes", len(msg.SystemOwnerChanges))
			for _, change := range msg.SystemOwnerChanges {
				err := game.SetSystemOwner(m.galaxy, change.SystemId, change.Owner)
				if err != nil {
					debugLog.Printf("Error setting system owner: %v", err)
				} else {
					debugLog.Printf("System %d owner changed to %s",
						change.SystemId, change.Owner)
					m.gameLogger.AddSystemOwnerChange(change)
				}
			}
		}

		if len(msg.GesUpdates) > 0 {
			for _, update := range msg.GesUpdates {
				if update.Owner == m.username {
					debugLog.Printf("Received GES update: %d", update.Amount)
					m.gesAmount = update.Amount
				}
			}
		}

		if len(msg.FleetCreations) > 0 {
			for _, creation := range msg.FleetCreations {
				err := game.ProcessFleetCreation(m.galaxy, creation)
				if err != nil {
					debugLog.Printf("Error processing fleet creation: %v", err)
				} else {
					debugLog.Printf("Fleet created with ID %d in system %d by %s",
						creation.FleetId, creation.SystemId, creation.Owner)
					m.gameLogger.AddFleetCreation(creation)
				}
			}
		}

		if msg.Victory != nil {
			debugLog.Printf("Game victory: %s has won!", msg.Victory.Winner)
			m.gameLogger.AddVictory(msg.Victory)
			ui.UpdateLogWindow(m.logWindow, m.gameLogger)

			time.Sleep(2 * time.Second)

			// Reset to initial state, preserving only logs and connections
			m.username = ""
			m.started = false
			m.joined = false
			m.playerCount = 0
			m.players = make(map[string]*pb.Player)
			m.galaxy = nil
			m.command = ""
			m.gesAmount = 0
			m.selectedX = 0
			m.selectedY = 0
			m.inspector = nil
			ui.ResetEnemyColors()

			m.gameLogger.AddSystemMessage("Game has been reset. Enter your name to play again.")
			ui.UpdateLogWindow(m.logWindow, m.gameLogger)

			return m, nil
		}
		ui.UpdateLogWindow(m.logWindow, m.gameLogger)

		return m, nil
	case game.ErrorMessage:
		debugLog.Printf("Error from server: %s", msg.Content)
		m.command = fmt.Sprintf("ERROR: %s", msg.Content)
		m.gameLogger.AddSystemMessage(fmt.Sprintf("Error: %s", msg.Content))
		ui.UpdateLogWindow(m.logWindow, m.gameLogger)
		return m, nil
	}
	return m, nil
}
