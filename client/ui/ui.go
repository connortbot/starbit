package ui

import (
	"fmt"
	"strings"
	"sort"

	pb "starbit/proto"
)

func RenderPlayerList(username string, players map[string]*pb.Player) string {
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
			s.WriteString(fmt.Sprintf("│ ★ %-11s \n", name))
		} else {
			s.WriteString(fmt.Sprintf("│   %-11s \n", name))
		}
	}
	s.WriteString("╰─────────────────╯\n")
	s.WriteString("\nPress Ctrl+C to quit\n")
	return s.String()
}