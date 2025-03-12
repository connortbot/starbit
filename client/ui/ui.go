package ui

import (
	"fmt"
	"sort"
	"strings"

	pb "starbit/proto"

	"github.com/charmbracelet/lipgloss"
)

type TitleAlignment int

// UI Constants for layout
const (
	InspectorHeight  = 14
	InspectorWidth   = 60
	GalaxyBoxWidth   = 40
	CommandLineWidth = InspectorWidth + GalaxyBoxWidth + 2
	PlayerBoxWidth   = 30
	PlayerBoxHeight  = 15
	LogBoxWidth      = 60
	LogBoxHeight     = 15
)

var (
	redStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	boldStyle  = lipgloss.NewStyle().Bold(true)
)

func RenderCommandLine(command string) string {
	var s strings.Builder
	s.WriteString(renderBoxTop(CommandLineWidth, "Command", TitleLeft) + "\n")
	s.WriteString(padLine("> "+command, CommandLineWidth, true) + "\n")
	s.WriteString(renderBoxBottom(CommandLineWidth))
	return s.String()
}

func RenderGuideLabel(text string) string {
	insideWidth := lipgloss.Width(text) + 4
	var s strings.Builder
	s.WriteString(renderBoxTop(insideWidth, "", TitleLeft) + "\n")
	s.WriteString(padLine(text, insideWidth, true) + "\n")
	s.WriteString(renderBoxBottom(insideWidth))
	return s.String()
}

func RenderHelpFooter() string {
	return sideBySideBoxes(
		2,
		RenderGuideLabel("Cmd: Shift+C"),
		RenderGuideLabel("Inspect: Shift+I"),
		RenderGuideLabel("Explore: Shift+E"),
		RenderGuideLabel("Quit: Ctrl+C"),
	)
}

func GenerateInspectContent(width int, system *pb.System) string {
	// quick info boxes
	idInfo := wrapInBox(fmt.Sprintf("ID: %d", system.Id), 10, 0, "", TitleCenter)
	locationInfo := wrapInBox(fmt.Sprintf("Location: %d, %d", system.X, system.Y), 20, 0, "", TitleCenter)
	owner := "None"
	if system.Owner != "none" {
		owner = system.Owner
	}
	ownerWidth := lipgloss.Width(fmt.Sprintf("Owner: %s", owner))
	ownerInfo := wrapInBox(fmt.Sprintf("Owner: %s", owner), ownerWidth+4, 0, "", TitleCenter)

	infoSection := sideBySideBoxes(1, idInfo, locationInfo, ownerInfo)

	var content string
	if len(system.Fleets) > 0 {
		fleetListBoxes := []string{}
		for _, fleet := range system.Fleets {
			fleetListBoxes = append(fleetListBoxes, RenderFleet(fleet, 50))
		}
		fleetContent := listBoxes(1, fleetListBoxes...)
		content = listBoxes(1, infoSection, fleetContent)
	} else {
		content = infoSection
	}

	return content
}

func NewInspectWindow(width int, system *pb.System) *ScrollingViewport {
	content := GenerateInspectContent(width, system)
	return NewScrollingViewport(
		content,
		calculateStretchBoxWidth(content, "Inspector", width),
		InspectorHeight,
		"Inspector",
		TitleCenter,
	)
}

func RenderInspectWindow(inspector *ScrollingViewport) string {
	return inspector.Render()
}

func RenderPlayerBox(started bool, username string, players map[string]*pb.Player) string {
	var s strings.Builder
	s.WriteString(renderBoxTop(PlayerBoxWidth, "Players", TitleCenter) + "\n")
	s.WriteString(padLine(fmt.Sprintf("Players: %d", len(players)), PlayerBoxWidth, false) + "\n")

	var names []string
	for _, player := range players {
		names = append(names, player.Name)
	}
	sort.Strings(names)

	for _, name := range names {
		if name == username {
			s.WriteString(padLine(fmt.Sprintf("★ %s", name), PlayerBoxWidth, false) + "\n")
		} else {
			s.WriteString(padLine(fmt.Sprintf("  %s", name), PlayerBoxWidth, false) + "\n")
		}
	}

	remainingLines := PlayerBoxHeight - 3 - len(names)
	for i := 0; i < remainingLines; i++ {
		s.WriteString(padLine("", PlayerBoxWidth, false) + "\n")
	}

	s.WriteString(renderMidline(PlayerBoxWidth))

	if started {
		s.WriteString(padLine(greenStyle.Render("     Ready     "), PlayerBoxWidth, false) + "\n")
	} else {
		s.WriteString(padLine(redStyle.Render("   Not Ready   "), PlayerBoxWidth, false) + "\n")
	}

	s.WriteString(renderBoxBottom(PlayerBoxWidth))

	return s.String()
}

func RenderGameScreen(
	username string,
	players map[string]*pb.Player,
	started bool,
	galaxy *pb.GalaxyState,
	command string,
	inspector *ScrollingViewport,
	logWindow *ScrollingViewport,
	selectedX int32,
	selectedY int32,
	selectedSystem *pb.System,
	controlMode string,
	gesAmount int32,
) string {
	var s strings.Builder

	playerBox := RenderPlayerBox(started, username, players)

	if logWindow != nil {
		s.WriteString(sideBySideBoxes(2, playerBox, logWindow.Render()))
	} else {
		s.WriteString(playerBox + "\n")
	}

	if started && galaxy != nil {
		galaxyContent := RenderGalaxy(galaxy, username, selectedX, selectedY)
		boxedGalaxyContent := wrapInBox(galaxyContent, GalaxyBoxWidth, 0, "Galaxy", TitleCenter)
		inspector.UpdateContent(GenerateInspectContent(InspectorWidth, selectedSystem))
		inspectWindow := RenderInspectWindow(inspector)
		s.WriteString(sideBySideBoxes(2, boxedGalaxyContent, inspectWindow))
		s.WriteString(RenderCommandLine(command))
		s.WriteString(RenderHelpFooter())
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("   Mode: %s    GES: %d", controlMode, gesAmount))
	}
	return s.String()
}

func RenderJoinScreen(username string, logWindow *ScrollingViewport, ipAddress string, tcpPort string, udpPort string, connected bool) string {
	var s strings.Builder

	usernameBox := strings.Builder{}
	usernameBox.WriteString(renderBoxTop(PlayerBoxWidth, "Username (N)", TitleLeft) + "\n")
	usernameBox.WriteString(padLine("> "+username, PlayerBoxWidth, false) + "\n")
	usernameBox.WriteString(renderBoxBottom(PlayerBoxWidth))

	ipBox := strings.Builder{}
	ipBox.WriteString(renderBoxTop(PlayerBoxWidth, "IP Address (I)", TitleLeft) + "\n")
	ipBox.WriteString(padLine("> "+ipAddress, PlayerBoxWidth, false) + "\n")
	ipBox.WriteString(renderBoxBottom(PlayerBoxWidth))

	tcpBox := strings.Builder{}
	tcpBox.WriteString(renderBoxTop(PlayerBoxWidth, "TCP Port (T)", TitleLeft) + "\n")
	tcpBox.WriteString(padLine("> "+tcpPort, PlayerBoxWidth, false) + "\n")
	tcpBox.WriteString(renderBoxBottom(PlayerBoxWidth))

	udpBox := strings.Builder{}
	udpBox.WriteString(renderBoxTop(PlayerBoxWidth, "UDP Port (U)", TitleLeft) + "\n")
	udpBox.WriteString(padLine("> "+udpPort, PlayerBoxWidth, false) + "\n")
	udpBox.WriteString(renderBoxBottom(PlayerBoxWidth))

	joinButton := strings.Builder{}
	joinBoxTitle := "Connect to Starbit"
	if connected {
		joinBoxTitle = "Join Game"
	}
	joinButton.WriteString(renderBoxTop(PlayerBoxWidth, joinBoxTitle, TitleCenter) + "\n")
	joinButton.WriteString(padLine("Press ENTER", PlayerBoxWidth, true) + "\n")
	joinButton.WriteString(renderBoxBottom(PlayerBoxWidth))

	inputSection := listBoxes(1, usernameBox.String(), ipBox.String(), tcpBox.String(), udpBox.String(), joinButton.String())

	if logWindow != nil {
		s.WriteString(sideBySideBoxes(2, inputSection, logWindow.Render()))
	} else {
		s.WriteString(inputSection + "\n")
	}

	return s.String()
}
