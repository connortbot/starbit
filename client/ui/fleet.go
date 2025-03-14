package ui

import (
	"fmt"
	"strings"

	pb "starbit/proto"

	"github.com/charmbracelet/lipgloss"
)

func renderHealthBar(current, max int32, width int) string {
	healthNumWidth := len(fmt.Sprintf("%d", current))
	barWidth := width - 6 - healthNumWidth // account for "HP: " and the health number and space

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
	idText := fmt.Sprintf("Fleet ID: %d", fleet.Id)
	s.WriteString(boldStyle.Render(idText) + "\n\n")

	healthBar := renderHealthBar(fleet.Health, 100, width-4)
	s.WriteString(fmt.Sprintf("HP: %s %d\n\n", healthBar, fleet.Health))

	ownerBox := fmt.Sprintf("Owner: %s", fleet.Owner)
	atkInfo := fmt.Sprintf("Attack: %d", fleet.Attack)
	s.WriteString(sideBySideBoxes(4, ownerBox, atkInfo))

	return wrapInBox(s.String(), width, 0, "Fleet", TitleCenter, nil)
}

func RenderFleetWithLocation(fleet *pb.Fleet, location int32, width int) string {
	effectiveWidth := width - 4 // -4 for left and right borders and minimal padding

	var fleetDisplay strings.Builder
	idLocationText := fmt.Sprintf("Fleet ID: %d (Location: %d)", fleet.Id, location)
	fleetDisplay.WriteString(boldStyle.Render(idLocationText) + "\n\n")

	healthBar := renderHealthBar(fleet.Health, 100, effectiveWidth)
	fleetDisplay.WriteString(fmt.Sprintf("HP: %s %d\n\n", healthBar, fleet.Health))

	ownerBox := fmt.Sprintf("Owner: %s", fleet.Owner)
	atkInfo := fmt.Sprintf("Attack: %d", fleet.Attack)
	fleetDisplay.WriteString(sideBySideBoxes(4, ownerBox, atkInfo))
	return wrapInBox(fleetDisplay.String(), width, 0, "Fleet", TitleCenter, nil)
}

func GenerateFleetListContent(ownedFleets []*pb.Fleet, fleetLocations map[int32]int32, width int) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("\n Total Fleets: %d\n", len(ownedFleets)))
	fleetBoxWidth := width - 4

	fleetListBoxes := []string{}
	for _, fleet := range ownedFleets {
		fleetDisplay := RenderFleetWithLocation(fleet, fleetLocations[fleet.Id], fleetBoxWidth)
		fleetListBoxes = append(fleetListBoxes, fleetDisplay)
	}
	content.WriteString(listBoxes(1, fleetListBoxes...))
	return content.String()
}

func NewFleetListWindow(ownedFleets []*pb.Fleet, fleetLocations map[int32]int32, width, height int) *ScrollingViewport {
	content := GenerateFleetListContent(ownedFleets, fleetLocations, width)
	return NewScrollingViewport(
		content,
		width,
		height,
		"Your Fleets",
		TitleCenter,
	)
}

func UpdateFleetListWindow(viewport *ScrollingViewport, ownedFleets []*pb.Fleet, fleetLocations map[int32]int32, width int) {
	content := GenerateFleetListContent(ownedFleets, fleetLocations, width)
	viewport.UpdateContent(content)
}

func RenderFleetListWindow(fleetList *ScrollingViewport, style *lipgloss.Style) string {
	return fleetList.Render(style)
}
