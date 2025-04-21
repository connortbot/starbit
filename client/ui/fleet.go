package ui

import (
	"fmt"
	"strings"

	pb "starbit/proto"

	"github.com/charmbracelet/lipgloss"
)

const (
	FLEET_MOVEMENT_COOLDOWN = 10
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

	ownerBox := fmt.Sprintf("Owner: %s", fleet.Owner)
	s.WriteString(sideBySideBoxes(2, boldStyle.Render(idText), ownerBox))
	s.WriteString("\n")	
	healthInfo := fmt.Sprintf("%d/%d", fleet.Health, fleet.MaxHealth)
	healthBar := renderHealthBar(fleet.Health, fleet.MaxHealth, width-(2 + len(healthInfo)))
	s.WriteString(fmt.Sprintf("HP: %s %s\n\n", healthBar, healthInfo))

	atkInfo := fmt.Sprintf("Atk: %d", fleet.Attack)
	exatkInfo := fmt.Sprintf("ExAtk: %d", fleet.Exattack)
	evasionInfo := fmt.Sprintf("Eva: %d%%", fleet.Evasion)
	armorInfo := fmt.Sprintf("Arm: %d%%", fleet.Armor)
	
	s.WriteString(sideBySideBoxes(2, atkInfo, exatkInfo, evasionInfo, armorInfo))

	return wrapInBox(s.String(), width, 0, "Fleet", TitleCenter, nil)
}

func RenderFleetWithLocation(fleet *pb.Fleet, location int32, width int, currentTickCount int32) string {

	var fleetDisplay strings.Builder	
	ownerBox := fmt.Sprintf("Owner: %s", fleet.Owner)
	idLocationText := fmt.Sprintf("Fleet ID: %d (Location: %d) %s", fleet.Id, location, ownerBox)
	fleetDisplay.WriteString(boldStyle.Render(idLocationText) + "\n\n")
	
	healthInfo := fmt.Sprintf("%d/%d", fleet.Health, fleet.MaxHealth)
	healthBar := renderHealthBar(fleet.Health, fleet.MaxHealth, width-(2 + len(healthInfo)))
	fleetDisplay.WriteString(fmt.Sprintf("HP: %s %s\n\n", healthBar, healthInfo))

	atkInfo := fmt.Sprintf("Atk: %d", fleet.Attack)
	exatkInfo := fmt.Sprintf("ExAtk: %d", fleet.Exattack)
	evasionInfo := fmt.Sprintf("Eva: %d%%", fleet.Evasion)
	armorInfo := fmt.Sprintf("Arm: %d%%", fleet.Armor)
	
	moveStatus := "READY"
	if fleet.LastMovedTick > (currentTickCount - FLEET_MOVEMENT_COOLDOWN) {
		ticksToWait := FLEET_MOVEMENT_COOLDOWN - (currentTickCount - fleet.LastMovedTick)
		moveStatus = fmt.Sprintf("%d sols", ticksToWait)
	}
	lastMovedText := fmt.Sprintf("Move: %s", moveStatus)
	fleetDisplay.WriteString(sideBySideBoxes(2, atkInfo, exatkInfo, evasionInfo, armorInfo, lastMovedText))
	
	return wrapInBox(fleetDisplay.String(), width, 0, "Fleet", TitleCenter, nil)
}

func GenerateFleetListContent(ownedFleets []*pb.Fleet, fleetLocations map[int32]int32, width int, currentTickCount int32) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("\n Total Fleets: %d\n", len(ownedFleets)))
	fleetBoxWidth := width - 4

	fleetListBoxes := []string{}
	for _, fleet := range ownedFleets {
		fleetDisplay := RenderFleetWithLocation(fleet, fleetLocations[fleet.Id], fleetBoxWidth, currentTickCount)
		fleetListBoxes = append(fleetListBoxes, fleetDisplay)
	}
	content.WriteString(listBoxes(1, fleetListBoxes...))
	return content.String()
}

func NewFleetListWindow(ownedFleets []*pb.Fleet, fleetLocations map[int32]int32, width, height int, currentTickCount int32) *ScrollingViewport {
	content := GenerateFleetListContent(ownedFleets, fleetLocations, width, currentTickCount)
	return NewScrollingViewport(
		content,
		width,
		height,
		"Your Fleets",
		TitleCenter,
	)
}

func UpdateFleetListWindow(viewport *ScrollingViewport, ownedFleets []*pb.Fleet, fleetLocations map[int32]int32, width int, currentTickCount int32) {
	content := GenerateFleetListContent(ownedFleets, fleetLocations, width, currentTickCount)
	viewport.UpdateContent(content)
}

func RenderFleetListWindow(fleetList *ScrollingViewport, style *lipgloss.Style) string {
	return fleetList.Render(style)
}
