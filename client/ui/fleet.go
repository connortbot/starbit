package ui

import (
	"fmt"
	"strings"

	pb "starbit/proto"
)

func renderHealthBar(current, max int32, width int) string {
	barWidth := width - 2 // account for brackets
	filled := int(float64(current) / float64(max) * float64(barWidth))
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}

	empty := barWidth - filled
	return fmt.Sprintf("%s%s",
		strings.Repeat("█", filled),
		strings.Repeat("░", empty))
}

func RenderFleet(fleet *pb.Fleet, width int) string {
	var s strings.Builder

	healthBar := renderHealthBar(fleet.Health, 100, width-10)
	s.WriteString(fmt.Sprintf("HP: %s %d\n\n", healthBar, fleet.Health))
	ownerBox := fmt.Sprintf("Owner: %s", fleet.Owner)
	atkInfo := fmt.Sprintf("Attack: %d", fleet.Attack)
	s.WriteString(sideBySideBoxes(4, ownerBox, atkInfo))

	return wrapInBox(s.String(), width, 0, "Fleet", TitleCenter)
}
