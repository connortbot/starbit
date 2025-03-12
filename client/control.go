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
		case "enter":
			if m.username != "" && !m.started {
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

				// log
				debugLog.Printf("Joined game successfully. Started: %v, Players: %d",
					m.started, m.playerCount)

				return m, nil
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
	case game.GameMsg:
		m.playerCount = msg.PlayerCount
		m.started = msg.Started
		m.players = msg.Players
		if msg.Galaxy != nil {
			m.galaxy = msg.Galaxy
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
					m.command = ""
				} else {
					debugLog.Printf("Command failed: %s", result.Message)
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
				}
			}
		}
		return m, nil
	case game.ErrorMessage:
		debugLog.Printf("Error from server: %s", msg.Content)
		m.command = fmt.Sprintf("ERROR: %s", msg.Content)
		return m, nil
	}
	return m, nil
}
