package main

import (
	"starbit/client/game"

	"starbit/client/ui"

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
				if err := m.client.SubscribeToTicks(m.username); err != nil {
					m.client.Close()
					m.err = err
					return m, nil
				}
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
				m.inspector = ui.NewInspectWindow(60, m.galaxy.Systems[0])
				m.controlMode = CommandMode
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
	case game.TickMsg:
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
			if m.controlMode == CommandMode {
				debugLog.Printf("Enter: %s", m.command)
			}
		case "backspace":
			if m.controlMode == CommandMode && len(m.command) > 0 {
				m.command = m.command[:len(m.command)-1]
			}
		default:
			if m.controlMode == CommandMode && len(msg.String()) == 1 {
				m.command += msg.String()
			}
		}
	case game.TickMsg:
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
