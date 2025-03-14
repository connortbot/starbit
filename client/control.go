package main

import (
	"log"
	"starbit/client/game"

	"starbit/client/ui"

	"time"

	pb "starbit/proto"

	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) HandleFirstScreen(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.firstScreen = false
		}
	}
	return m, nil
}

func (m model) HandleMenu(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.client != nil {
				m.client.Close()
			}
			return m, tea.Quit
		case "I", "shift+i":
			m.inputMode = IPMode
		case "N", "shift+n":
			m.inputMode = UsernameMode
		case "enter":
			if !m.connected {
				readyToConnect := m.ipAddress != "" && !m.started
				if !readyToConnect {
					m.gameLogger.AddSystemMessage("Please enter a IP address, TCP port, and UDP port to connect to the server.")
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					return m, nil
				}
				m.client.SetConnectionInfo(m.ipAddress, "50051")
				m.udpClient.SetConnectionInfo(m.ipAddress, "50052")

				m.client.Connect()
				if err := m.client.Connect(); err != nil {
					log.Printf("Error connecting TCP client: %v", err)
					m.gameLogger.AddSystemMessage(fmt.Sprintf("Error connecting to TCP client [%v] Double check your IP and TCP port!", err))
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					return m, nil
				}
				log.Printf("TCP client connected successfully")
				m.gameLogger.AddSystemMessage("TCP client connected successfully")
				ui.UpdateLogWindow(m.logWindow, m.gameLogger)

				if err := m.udpClient.Connect(); err != nil {
					log.Printf("Error connecting UDP client: %v", err)
					m.gameLogger.AddSystemMessage(fmt.Sprintf("Error connecting to UDP client [%v] Double check your IP and UDP port!", err))
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					return m, nil
				}
				log.Printf("UDP client connected successfully")
				m.gameLogger.AddSystemMessage("UDP client connected successfully")
				ui.UpdateLogWindow(m.logWindow, m.gameLogger)

				go listenForUDPTicks(m.udpClient, program)
				go listenForUDPErrors(m.udpClient, program)
				go listenForTCPUpdates(m.client, program)
				log.Printf("Listening for UDP ticks, errors, and TCP updates")

				m.connected = true
				return m, nil
			} else {
				log.Printf("Joining Game")
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
				m.fleetList = ui.NewFleetListWindow(m.ownedFleets, m.fleetLocations, ui.FleetListWidth, ui.FleetListHeight)
				m.controlMode = CommandMode

				m.gameLogger.AddSystemMessage(fmt.Sprintf("%s joined the game", m.username))
				m.logWindow = ui.NewLogWindow(m.gameLogger, ui.LogBoxWidth, ui.LogBoxHeight)

				if resp.Started {
					m.gameLogger.AddSystemMessage("Game started")
				}
				// log
				log.Printf("Joined game successfully. Started: %v, Players: %d",
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
			}
		default:
			if len(msg.String()) == 1 && !m.started {
				if m.inputMode == UsernameMode {
					m.username += msg.String()
				} else if m.inputMode == IPMode {
					m.ipAddress += msg.String()
				}
			}
		}
	case game.GameMsg:
		m.playerCount = msg.PlayerCount
		m.started = msg.Started
		m.players = msg.Players
		if msg.Galaxy != nil {
			m.galaxy = msg.Galaxy
			m.ownedFleets, m.fleetLocations = game.FindOwnedFleetsAndLocations(m.galaxy, m.username)
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
		case "F", "shift+f":
			m.controlMode = FleetListMode
		case "up":
			if m.controlMode == InspectMode {
				m.inspector.ScrollUp()
			} else if m.controlMode == FleetListMode && m.fleetList != nil {
				m.fleetList.ScrollUp()
			} else if m.controlMode == ExploreMode {
				if m.selectedY > 0 {
					m.selectedY--
				}
				m.inspector.ScrollToTop()
			}
		case "down":
			if m.controlMode == InspectMode {
				m.inspector.ScrollDown()
			} else if m.controlMode == FleetListMode && m.fleetList != nil {
				m.fleetList.ScrollDown()
			} else if m.controlMode == ExploreMode {
				if m.galaxy != nil && m.selectedY < m.galaxy.Height-1 {
					m.selectedY++
				}
				m.inspector.ScrollToTop()
			}
		case "shift+up":
			if m.controlMode == InspectMode {
				m.inspector.ScrollToTop()
			} else if m.controlMode == FleetListMode && m.fleetList != nil {
				m.fleetList.ScrollToTop()
			}
		case "shift+down":
			if m.controlMode == InspectMode {
				m.inspector.ScrollToBottom()
			} else if m.controlMode == FleetListMode && m.fleetList != nil {
				m.fleetList.ScrollToBottom()
			}
		case "left":
			if m.controlMode == ExploreMode {
				if m.selectedX > 0 {
					m.selectedX--
				}
				m.inspector.ScrollToTop()
			}
		case "right":
			if m.controlMode == ExploreMode {
				if m.galaxy != nil && m.selectedX < m.galaxy.Width-1 {
					m.selectedX++
				}
				m.inspector.ScrollToTop()
			}
		case "enter":
			if m.controlMode == CommandMode && m.command != "" {
				log.Printf("Processing command: %s", m.command)
				result := game.ParseCommand(m.udpClient, m.command)
				if result.Success {
					log.Printf("Command successful: %s", result.Message)
					m.gameLogger.AddCommand(m.username, m.command, true)
					ui.UpdateLogWindow(m.logWindow, m.gameLogger)
					m.command = ""
				} else {
					log.Printf("Command failed: %s", result.Message)
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
			m.ownedFleets, m.fleetLocations = game.FindOwnedFleetsAndLocations(m.galaxy, m.username)
			if m.fleetList != nil {
				ui.UpdateFleetListWindow(m.fleetList, m.ownedFleets, m.fleetLocations, ui.FleetListWidth)
			}
		}
		return m, nil
	case *pb.TickMsg:
		log.Printf("UDP Tick: %s", string(msg.Message))
		if len(msg.FleetMovements) > 0 {
			log.Printf("Received %d fleet movements", len(msg.FleetMovements))
			for _, movement := range msg.FleetMovements {
				err := game.MoveFleet(m.galaxy, movement)
				if err != nil {
					log.Printf("Error processing fleet movement: %v", err)
				} else {
					log.Printf("Fleet %d moved from system %d to system %d",
						movement.FleetId, movement.FromSystemId, movement.ToSystemId)
					m.gameLogger.AddFleetMovement(movement)
					fleet := game.GetFleet(m.galaxy, movement.ToSystemId, movement.FleetId)
					if fleet.Owner == m.username {
						m.fleetLocations[fleet.Id] = movement.ToSystemId
						if m.fleetList != nil {
							ui.UpdateFleetListWindow(m.fleetList, m.ownedFleets, m.fleetLocations, ui.FleetListWidth)
						}
					}
				}
			}
		}

		if len(msg.HealthUpdates) > 0 {
			log.Printf("Received %d health updates", len(msg.HealthUpdates))
			for _, update := range msg.HealthUpdates {
				err := game.ApplyHealthUpdate(m.galaxy, update)
				if err != nil {
					log.Printf("Error applying health update: %v", err)
				} else {
					log.Printf("Fleet %d health updated to %d in system %d",
						update.FleetId, update.Health, update.SystemId)
					m.gameLogger.AddHealthUpdate(update)
					if m.fleetList != nil {
						ui.UpdateFleetListWindow(m.fleetList, m.ownedFleets, m.fleetLocations, ui.FleetListWidth)
					}
				}
			}
		}

		if len(msg.FleetDestroyed) > 0 {
			log.Printf("Received %d destroyed fleets", len(msg.FleetDestroyed))
			for _, destroyed := range msg.FleetDestroyed {
				err := game.ProcessFleetDestroyed(m.galaxy, destroyed)
				if err != nil {
					log.Printf("Error processing destroyed fleet: %v", err)
				} else {
					log.Printf("Fleet %d was destroyed in system %d",
						destroyed.FleetId, destroyed.SystemId)
					m.gameLogger.AddFleetDestroyed(destroyed)
					m.ownedFleets = game.RemoveFromFleetArray(m.ownedFleets, destroyed.FleetId)
					delete(m.fleetLocations, destroyed.FleetId)
					log.Printf("Owned fleets: %v", m.ownedFleets)
					if m.fleetList != nil {
						ui.UpdateFleetListWindow(m.fleetList, m.ownedFleets, m.fleetLocations, ui.FleetListWidth)
					}
				}
			}
		}

		if len(msg.SystemOwnerChanges) > 0 {
			log.Printf("Received %d system owner changes", len(msg.SystemOwnerChanges))
			for _, change := range msg.SystemOwnerChanges {
				err := game.SetSystemOwner(m.galaxy, change.SystemId, change.Owner)
				if err != nil {
					log.Printf("Error setting system owner: %v", err)
				} else {
					log.Printf("System %d owner changed to %s",
						change.SystemId, change.Owner)
					m.gameLogger.AddSystemOwnerChange(change)
				}
			}
		}

		if len(msg.GesUpdates) > 0 {
			for _, update := range msg.GesUpdates {
				if update.Owner == m.username {
					log.Printf("Received GES update: %d", update.Amount)
					m.gesAmount = update.Amount
				}
			}
		}

		if len(msg.FleetCreations) > 0 {
			for _, creation := range msg.FleetCreations {
				err := game.ProcessFleetCreation(m.galaxy, creation)
				if err != nil {
					log.Printf("Error processing fleet creation: %v", err)
				} else {
					log.Printf("Fleet created with ID %d in system %d by %s",
						creation.FleetId, creation.SystemId, creation.Owner)
					m.gameLogger.AddFleetCreation(creation)

					if creation.Owner == m.username {
						fleet := game.GetFleet(m.galaxy, creation.SystemId, creation.FleetId)
						if fleet != nil {
							m.ownedFleets = append(m.ownedFleets, fleet)
							m.fleetLocations[fleet.Id] = creation.SystemId
							log.Printf("Owned fleets: %v", m.ownedFleets)
							if m.fleetList != nil {
								ui.UpdateFleetListWindow(m.fleetList, m.ownedFleets, m.fleetLocations, ui.FleetListWidth)
							}
						}
					}
				}
			}
		}

		if msg.Victory != nil {
			log.Printf("Game victory: %s has won!", msg.Victory.Winner)
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
			m.fleetList = nil
			m.fleetLocations = make(map[int32]int32)
			ui.ResetEnemyColors()

			m.gameLogger.AddSystemMessage("Game has been reset. Enter your name to play again.")
			ui.UpdateLogWindow(m.logWindow, m.gameLogger)

			return m, nil
		}
		ui.UpdateLogWindow(m.logWindow, m.gameLogger)

		return m, nil
	case game.ErrorMessage:
		log.Printf("Error from server: %s", msg.Content)
		m.command = fmt.Sprintf("ERROR: %s", msg.Content)
		m.gameLogger.AddSystemMessage(fmt.Sprintf("Error: %s", msg.Content))
		ui.UpdateLogWindow(m.logWindow, m.gameLogger)
		return m, nil
	}
	return m, nil
}
