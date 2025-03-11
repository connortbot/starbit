package ui

import (
	"strings"

	pb "starbit/proto"

	"github.com/charmbracelet/lipgloss"
)

var (
	systemBlueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0000FF")).
			Background(lipgloss.Color("#000066"))

	systemGreyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")).
			Background(lipgloss.Color("#222222"))

	systemRedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#660000"))
)

func RenderGalaxy(galaxy *pb.GalaxyState, username string) string {
	var s strings.Builder
	s.WriteString("Galaxy Map:\n")

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
		for x := int32(0); x <= maxX; x++ {
			system := grid[y][x]
			if system == nil {
				s.WriteString(systemGreyStyle.Render("·"))
				continue
			}

			switch {
			case system.Owner == username:
				s.WriteString(systemBlueStyle.Render("■"))
			case system.Owner == "none":
				s.WriteString(systemGreyStyle.Render("□"))
			default:
				s.WriteString(systemRedStyle.Render("■"))
			}
			s.WriteString(" ")
		}
		s.WriteString("\n")
	}

	return s.String()
}
