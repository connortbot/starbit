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
// 180 x 45
const (
	PlayerBoxWidth  = 30
	PlayerBoxHeight = 15
	LogBoxWidth     = 90
	LogBoxHeight    = 15

	GalaxyBoxWidth  = 40
	InspectorHeight = 14
	InspectorWidth  = PlayerBoxWidth + LogBoxWidth - GalaxyBoxWidth

	CommandLineWidth = InspectorWidth + GalaxyBoxWidth + 1
	FleetListWidth   = 70
	FleetListHeight  = LogBoxHeight + InspectorHeight + 3 + 2
)

var (
	redStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	boldStyle  = lipgloss.NewStyle().Bold(true)
)

func RenderCommandLine(command string, style *lipgloss.Style) string {
	var s strings.Builder
	s.WriteString(renderBoxTop(CommandLineWidth, "Command", TitleLeft, style) + "\n")
	s.WriteString(padLine("> "+command, CommandLineWidth, true, style) + "\n")
	s.WriteString(renderBoxBottom(CommandLineWidth, style))
	return s.String()
}

func RenderGuideLabel(text string) string {
	insideWidth := lipgloss.Width(text) + 4
	var s strings.Builder
	s.WriteString(renderBoxTop(insideWidth, "", TitleLeft, nil) + "\n")
	s.WriteString(padLine(text, insideWidth, true, nil) + "\n")
	s.WriteString(renderBoxBottom(insideWidth, nil))
	return s.String()
}

func RenderHelpFooter() string {
	cmdGuide := RenderGuideLabel("Cmd: Shift+C")
	inspectGuide := RenderGuideLabel("Inspect: Shift+I")
	exploreGuide := RenderGuideLabel("Explore: Shift+E")
	fleetListGuide := RenderGuideLabel("Fleets: Shift+F")
	quitGuide := RenderGuideLabel("Quit: Ctrl+C")

	return sideBySideBoxes(2, cmdGuide, inspectGuide, exploreGuide, fleetListGuide, quitGuide)
}

func GenerateInspectContent(width int, system *pb.System) string {
	// quick info boxes
	idInfo := wrapInBox(fmt.Sprintf("ID: %d", system.Id), 10, 0, "", TitleCenter, nil)
	locationInfo := wrapInBox(fmt.Sprintf("Location: %d, %d", system.X, system.Y), 20, 0, "", TitleCenter, nil)
	owner := "None"
	if system.Owner != "none" {
		owner = system.Owner
	}
	ownerWidth := lipgloss.Width(fmt.Sprintf("Owner: %s", owner))
	ownerInfo := wrapInBox(fmt.Sprintf("Owner: %s", owner), ownerWidth+4, 0, "", TitleCenter, nil)

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

func RenderInspectWindow(inspector *ScrollingViewport, style *lipgloss.Style) string {
	return inspector.Render(style)
}

func RenderPlayerBox(started bool, username string, players map[string]*pb.Player) string {
	var s strings.Builder
	s.WriteString(renderBoxTop(PlayerBoxWidth, "Players", TitleCenter, nil) + "\n")
	s.WriteString(padLine(fmt.Sprintf("Players: %d", len(players)), PlayerBoxWidth, false, nil) + "\n")

	var names []string
	for _, player := range players {
		names = append(names, player.Name)
	}
	sort.Strings(names)

	for _, name := range names {
		if name == username {
			s.WriteString(padLine(fmt.Sprintf("★ %s", name), PlayerBoxWidth, false, nil) + "\n")
		} else {
			s.WriteString(padLine(fmt.Sprintf("  %s", name), PlayerBoxWidth, false, nil) + "\n")
		}
	}

	remainingLines := PlayerBoxHeight - 3 - len(names)
	for i := 0; i < remainingLines; i++ {
		s.WriteString(padLine("", PlayerBoxWidth, false, nil) + "\n")
	}

	s.WriteString(renderMidline(PlayerBoxWidth, nil))

	if started {
		s.WriteString(padLine(greenStyle.Render("     Ready     "), PlayerBoxWidth, false, nil) + "\n")
	} else {
		s.WriteString(padLine(redStyle.Render("   Not Ready   "), PlayerBoxWidth, false, nil) + "\n")
	}

	s.WriteString(renderBoxBottom(PlayerBoxWidth, nil))

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
	fleetList *ScrollingViewport,
	selectedX int32,
	selectedY int32,
	selectedSystem *pb.System,
	controlMode string,
	gesAmount int32,
	currentTickCount int32,
) string {
	var s strings.Builder

	playerBox := RenderPlayerBox(started, username, players)
	var infoSection string
	if logWindow != nil {
		infoSection = sideBySideBoxes(1, playerBox, logWindow.Render(nil))
	} else {
		infoSection = playerBox + "\n"
	}

	if started && galaxy != nil {
		galaxyContent := RenderGalaxy(galaxy, username, selectedX, selectedY)
		var boxedGalaxyStyle *lipgloss.Style
		var boxedCommandStyle *lipgloss.Style
		var boxedInspectorStyle *lipgloss.Style
		var boxedFleetListStyle *lipgloss.Style
		if controlMode == "Explore" {
			boxedGalaxyStyle = &greenStyle
		} else if controlMode == "Inspect" {
			boxedInspectorStyle = &greenStyle
		} else if controlMode == "Command" {
			boxedCommandStyle = &greenStyle
		} else if controlMode == "FleetList" {
			boxedFleetListStyle = &greenStyle
		}

		boxedGalaxyContent := wrapInBox(galaxyContent, GalaxyBoxWidth, 0, "Galaxy", TitleCenter, boxedGalaxyStyle)
		inspector.UpdateContent(GenerateInspectContent(InspectorWidth, selectedSystem))
		inspectWindow := inspector.Render(boxedInspectorStyle)
		fleetListWindow := fleetList.Render(boxedFleetListStyle)

		gameplayViewportContent := sideBySideBoxes(1, boxedGalaxyContent, inspectWindow)
		gameContent := listBoxes(0,
			gameplayViewportContent,
			RenderCommandLine(command, boxedCommandStyle)+"\n",
			RenderHelpFooter()) + "\n"
		leftContent := listBoxes(0, infoSection, gameContent, fmt.Sprintf("   Mode: %s    GES: %d    Sol: %d", controlMode, gesAmount, currentTickCount))
		s.WriteString(
			sideBySideBoxes(1,
				leftContent,
				fleetListWindow,
			),
		)
	} else {
		s.WriteString(infoSection)
	}
	return s.String()
}

func RenderJoinScreen(username string, logWindow *ScrollingViewport, ipAddress string, connected bool, inputMode string) string {
	var s strings.Builder

	var usernameBoxStyle *lipgloss.Style
	var ipBoxStyle *lipgloss.Style

	if inputMode == "Username" {
		usernameBoxStyle = &greenStyle
	} else if inputMode == "IP" {
		ipBoxStyle = &greenStyle
	}

	usernameContent := "> " + username
	usernameBox := wrapInBox(
		usernameContent,
		PlayerBoxWidth,
		0,
		"Username (Shift+N)",
		TitleLeft,
		usernameBoxStyle,
	)

	ipBox := wrapInBox(
		"> "+ipAddress,
		PlayerBoxWidth,
		0,
		"IP Address (Shift+I)",
		TitleLeft,
		ipBoxStyle,
	)

	joinBoxTitle := "Connect to Starbit"
	if connected {
		joinBoxTitle = "Join Game"
	}

	joinButton := wrapInBox(
		"Press ENTER",
		PlayerBoxWidth,
		0,
		joinBoxTitle,
		TitleCenter,
		nil,
	)

	inputSection := listBoxes(1, usernameBox, ipBox, joinButton)

	if logWindow != nil {
		s.WriteString(sideBySideBoxes(2, inputSection, logWindow.Render(nil)))
	} else {
		s.WriteString(inputSection + "\n")
	}

	return s.String()
}

func RenderFirstScreen() string {
	width := GalaxyBoxWidth + InspectorWidth + FleetListWidth + 3
	height := LogBoxHeight + InspectorHeight + 12

	content := strings.Repeat("\n", height-2) + "For the best experience, expand your terminal window until you can see the edges of this box.\nPRESS ENTER WHEN READY :)"
	box := wrapInBox(content, width, height, "Starbit", TitleLeft, &greenStyle)
	return box
}
