package ui

import (
	"fmt"
	"strings"

	pb "starbit/proto"

	"github.com/charmbracelet/lipgloss"
)

var (
	systemBlueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0088FF"))
	systemGreyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777"))
	systemRedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))
)

func RenderGalaxy(galaxy *pb.GalaxyState, username string, selectedX, selectedY int32) string {
	var s strings.Builder
	rows := make(map[int32][]*pb.System)
	var maxY, maxX int32
	for _, system := range galaxy.Systems {
		rows[system.Y] = append(rows[system.Y], system)
		if system.Y > maxY {
			maxY = system.Y
		}
		if system.X > maxX {
			maxX = system.X
		}
	}

	grid := make([][]*pb.System, maxY+1)
	for y := range grid {
		grid[y] = make([]*pb.System, maxX+1)
	}
	for _, system := range galaxy.Systems {
		grid[system.Y][system.X] = system
	}

	for y := int32(0); y <= maxY; y++ {
		// First line of systems - shows the symbol
		for x := int32(0); x <= maxX; x++ {
			system := grid[y][x]

			if system == nil {
				s.WriteString(systemGreyStyle.Render("   ⋅   "))
				continue
			}

			var symbol string
			if x == selectedX && y == selectedY {
				symbol = "   ✦   " // selected system
			} else {
				switch {
				case system.Owner == username:
					symbol = "   ●   " // player's system
				case system.Owner == "none":
					symbol = "   ○   " // unclaimed system
				default:
					symbol = "   ◆   " // enemy system
				}
			}

			switch {
			case x == selectedX && y == selectedY:
				s.WriteString(selectedStyle.Render(symbol))
			case system.Owner == username:
				s.WriteString(systemBlueStyle.Render(symbol))
			case system.Owner == "none":
				s.WriteString(systemGreyStyle.Render(symbol))
			default:
				s.WriteString(systemRedStyle.Render(symbol))
			}
		}
		s.WriteString("\n")

		// Second line - shows the system ID
		for x := int32(0); x <= maxX; x++ {
			system := grid[y][x]

			if system == nil {
				s.WriteString(systemGreyStyle.Render("       "))
				continue
			}

			// Create ID label with consistent width
			var idStr string
			if system.Id < 10 {
				idStr = " ID:0" + fmt.Sprintf("%d", system.Id) + " " // Pad single digit with leading zero
			} else {
				idStr = " ID:" + fmt.Sprintf("%d", system.Id) + " "
			}

			switch {
			case x == selectedX && y == selectedY:
				s.WriteString(selectedStyle.Render(idStr))
			case system.Owner == username:
				s.WriteString(systemBlueStyle.Render(idStr))
			case system.Owner == "none":
				s.WriteString(systemGreyStyle.Render(idStr))
			default:
				s.WriteString(systemRedStyle.Render(idStr))
			}
		}
		s.WriteString("\n")

		// Add an empty line between rows for more spacing
		s.WriteString("\n")
	}

	return s.String()
}
