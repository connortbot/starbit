package ui

import (
	"fmt"
	"sort"
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
	systemPurpleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AA55FF"))
	systemOrangeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF9900"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	enemyColorMap map[string]lipgloss.Style
)

func ResetEnemyColors() {
	enemyColorMap = nil
}

// InitializeEnemyColors creates a consistent color mapping for enemy players
// this should be called once when the galaxy state is first received
func InitializeEnemyColors(galaxy *pb.GalaxyState, myUsername string) {
	enemyColorMap = make(map[string]lipgloss.Style)

	uniqueOwners := make(map[string]bool)
	for _, system := range galaxy.Systems {
		if system.Owner != myUsername && system.Owner != "none" {
			uniqueOwners[system.Owner] = true
		}
	}

	sortedOwners := make([]string, 0, len(uniqueOwners))
	for owner := range uniqueOwners {
		sortedOwners = append(sortedOwners, owner)
	}
	sort.Strings(sortedOwners)

	// available enemy colors
	enemyColors := []lipgloss.Style{systemRedStyle, systemPurpleStyle, systemOrangeStyle}

	for i, owner := range sortedOwners {
		if i < len(enemyColors) {
			enemyColorMap[owner] = enemyColors[i]
		} else {
			enemyColorMap[owner] = systemRedStyle
		}
	}
}

func RenderGalaxy(galaxy *pb.GalaxyState, username string, selectedX, selectedY int32) string {
	if enemyColorMap == nil {
		InitializeEnemyColors(galaxy, username)
	}

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
				if style, ok := enemyColorMap[system.Owner]; ok {
					s.WriteString(style.Render(symbol))
				} else {
					s.WriteString(systemRedStyle.Render(symbol))
				}
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
				if style, ok := enemyColorMap[system.Owner]; ok {
					s.WriteString(style.Render(idStr))
				} else {
					s.WriteString(systemRedStyle.Render(idStr))
				}
			}
		}
		s.WriteString("\n")

		// Add an empty line between rows for more spacing
		s.WriteString("\n")
	}

	return s.String()
}
