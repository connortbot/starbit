package ui

import (
	"fmt"
	"sort"
	"strings"

	pb "starbit/proto"

	"github.com/charmbracelet/lipgloss"
)

var (
	redStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
)

func RenderPlayerList(username string, players map[string]*pb.Player, started bool, galaxy *pb.GalaxyState) string {
	var s strings.Builder
	s.WriteString("╭──── Starbit ────╮\n")
	s.WriteString(fmt.Sprintf("│ Players: %d/2    │\n", len(players)))
	s.WriteString("├─────────────────┤\n")
	var names []string
	for _, player := range players {
		names = append(names, player.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		if name == username {
			s.WriteString(fmt.Sprintf("│ ★ %-11s │\n", name))
		} else {
			s.WriteString(fmt.Sprintf("│   %-11s │\n", name))
		}
	}
	s.WriteString("├─────────────────┤\n")
	if len(players) == 2 {
		s.WriteString(fmt.Sprintf("│ %s │\n", greenStyle.Render("     Ready     ")))
	} else {
		s.WriteString(fmt.Sprintf("│ %s │\n", redStyle.Render("   Not Ready   ")))
	}
	s.WriteString("╰─────────────────╯\n\n")

	if started && galaxy != nil {
		s.WriteString(RenderGalaxy(galaxy, username))
		s.WriteString("\n")
	}

	s.WriteString("Press Ctrl+C to quit\n")
	return s.String()
}
