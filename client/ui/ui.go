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

func RenderCommandLine(command string, isActive bool) string {
	var style *lipgloss.Style
	if isActive {
		s := greenStyle
		style = &s
	}

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

func RenderInspectWindow(inspector *ScrollingViewport) string {
	return inspector.Render()
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
		var galaxyStyle, inspectorStyle *lipgloss.Style

		if controlMode == "Explore" {
			s := greenStyle
			galaxyStyle = &s
		} else if controlMode == "Inspect" {
			s := greenStyle
			inspectorStyle = &s
		}

		galaxyContent := RenderGalaxy(galaxy, username, selectedX, selectedY)
		boxedGalaxyContent := wrapInBox(galaxyContent, GalaxyBoxWidth, 0, "Galaxy", TitleCenter, galaxyStyle)

		if inspectorStyle != nil {
			inspector.SetStyle(inspectorStyle)
		} else {
			inspector.SetStyle(nil)
		}

		inspector.UpdateContent(GenerateInspectContent(InspectorWidth, selectedSystem))
		inspectWindow := RenderInspectWindow(inspector)
		s.WriteString(sideBySideBoxes(2, boxedGalaxyContent, inspectWindow))

		isCommandActive := controlMode == "Command"
		s.WriteString(RenderCommandLine(command, isCommandActive))

		s.WriteString(RenderHelpFooter())
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("   Mode: %s    GES: %d", controlMode, gesAmount))
	}
	return s.String()
}

func RenderJoinScreen(username string, logWindow *ScrollingViewport, ipAddress string, connected bool, inputMode string) string {
	var s strings.Builder

	var usernameStyle, ipStyle *lipgloss.Style

	if inputMode == "Username" {
		style := greenStyle
		usernameStyle = &style
	}

	if inputMode == "IP" {
		style := greenStyle
		ipStyle = &style
	}

	usernameBox := wrapInBox(
		"> "+username,
		PlayerBoxWidth,
		0,
		"Username (N)",
		TitleLeft,
		usernameStyle,
	)

	ipBox := wrapInBox(
		"> "+ipAddress,
		PlayerBoxWidth,
		0,
		"IP Address (I)",
		TitleLeft,
		ipStyle,
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
		s.WriteString(sideBySideBoxes(2, inputSection, logWindow.Render()))
	} else {
		s.WriteString(inputSection + "\n")
	}

	return s.String()
}
