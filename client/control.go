package main

import (
	"starbit/client/game"

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
		case "enter":
			debugLog.Printf("Enter: %s", m.command)
		case "backspace":
			if len(m.command) > 0 {
				m.command = m.command[:len(m.command)-1]
			}
		default:
			if len(msg.String()) == 1 {
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
