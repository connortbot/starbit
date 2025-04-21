package ui

import (
	"fmt"
	"strings"
	"time"

	pb "starbit/proto"
)

type GameLogger struct {
	messages []string
	maxLogs  int
}

func NewGameLogger(maxLogs int) *GameLogger {
	return &GameLogger{
		messages: []string{},
		maxLogs:  maxLogs,
	}
}

func (l *GameLogger) AddCommand(username, command string, success bool) {
	timestamp := time.Now().Format("15:04:05")
	var message string
	if success {
		message = fmt.Sprintf("[%s] %s executed: %s", timestamp, username, command)
	} else {
		message = fmt.Sprintf("[%s] %s failed: %s", timestamp, username, command)
	}
	l.addMessage(message)
}

func (l *GameLogger) AddFleetMovement(movement *pb.FleetMovement) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] Fleet %d moved from system %d to system %d",
		timestamp, movement.FleetId, movement.FromSystemId, movement.ToSystemId)
	l.addMessage(message)
}

func (l *GameLogger) AddFleetCreation(creation *pb.FleetCreation) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] %s created a new fleet (ID: %d) in system %d",
		timestamp, creation.Owner, creation.FleetId, creation.SystemId)
	l.addMessage(message)
}

func (l *GameLogger) AddFleetUpdate(update *pb.FleetUpdate) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] %s updated fleet (ID: %d) in system %d", timestamp, update.Owner, update.FleetId, update.SystemId)
	l.addMessage(message)
}

func (l *GameLogger) AddHealthUpdate(update *pb.HealthUpdate) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] Fleet %d health is now %d in system %d",
		timestamp, update.FleetId, update.Health, update.SystemId)
	l.addMessage(message)
}

func (l *GameLogger) AddFleetDestroyed(destroyed *pb.FleetDestroyed) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] Fleet %d was destroyed in system %d",
		timestamp, destroyed.FleetId, destroyed.SystemId)
	l.addMessage(message)
}

func (l *GameLogger) AddSystemOwnerChange(change *pb.SystemOwnerChange) {
	timestamp := time.Now().Format("15:04:05")
	ownerName := change.Owner
	if ownerName == "none" {
		ownerName = "contested"
	}
	message := fmt.Sprintf("[%s] System %d is now owned by %s",
		timestamp, change.SystemId, ownerName)
	l.addMessage(message)
}

func (l *GameLogger) AddGESUpdate(update *pb.GESUpdate) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] %s now has %d GES",
		timestamp, update.Owner, update.Amount)
	l.addMessage(message)
}

func (l *GameLogger) AddVictory(victory *pb.GameVictory) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] GAME OVER: %s has conquered the galaxy!",
		timestamp, victory.Winner)
	l.addMessage(message)
}

func (l *GameLogger) AddSystemMessage(message string) {
	timestamp := time.Now().Format("15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s", timestamp, message)
	l.addMessage(formattedMessage)
}

func (l *GameLogger) addMessage(message string) {
	l.messages = append(l.messages, message)
	if len(l.messages) > l.maxLogs {
		l.messages = l.messages[len(l.messages)-l.maxLogs:]
	}
}

func (l *GameLogger) GetContent() string {
	var s strings.Builder
	for _, msg := range l.messages {
		s.WriteString(msg + "\n")
	}
	return s.String()
}

func NewLogWindow(logger *GameLogger, width, height int) *ScrollingViewport {
	content := logger.GetContent()
	return NewScrollingViewport(
		content,
		calculateStretchBoxWidth(content, "Logs", width),
		height,
		"Logs",
		TitleCenter,
	)
}

func UpdateLogWindow(viewport *ScrollingViewport, logger *GameLogger) {
	viewport.UpdateContent(logger.GetContent())
	viewport.ScrollToBottom()
}
