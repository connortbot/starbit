package game

import (
	"fmt"
	"strconv"
	"strings"

	"log"

	pb "starbit/proto"
)

type CommandResult struct {
	Success bool
	Message string
}

// data needed for parsing all commands
type CommandData struct {
	FleetLocations map[int32]int32
	GalaxyRef *pb.GalaxyState
}

func ParseCommand(client *UDPClient, input string, data CommandData) CommandResult {
	input = strings.TrimSpace(input)
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return CommandResult{
			Success: false,
			Message: "Empty command",
		}
	}

	cmdType := strings.ToLower(parts[0])

	switch cmdType {
	case "fm":
		return handleFleetMovement(client, parts, &data)
	case "fc":
		return handleFleetCreation(client, parts)
	case "fu":
		return handleFleetModification(client, parts, &data)
	default:
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Unknown command: %s", cmdType),
		}
	}
}

// fu <fleet_id> <ship_type: de|cr|ba|dr>
type ShipType string
const (
	Destroyer ShipType = "de"
	Cruiser ShipType = "cr"
	Battleship ShipType = "ba"
	Dreadnought ShipType = "dr"
)

func (s ShipType) IsValid() bool {
	switch s {
	case Destroyer, Cruiser, Battleship, Dreadnought:
		return true
	}
	return false
}

func handleFleetModification(client *UDPClient, parts []string, data *CommandData) CommandResult {
	if len(parts) != 3 {
		return CommandResult{
			Success: false,
	 		Message: "Invalid fleet creation command. Format: fu <fleet_id> <ship_type: de|cr|ba|dr",
		}
	}


	fleetID, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Invalid fleet ID: %s", parts[1]),
		}
	}
	systemID, exists := data.FleetLocations[int32(fleetID)]
	if !exists {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Unowned fleet ID: %d", fleetID),
		}
	}
	
	sourceSystem := data.GalaxyRef.Systems[systemID]
	var fleet *pb.Fleet
	var fleet2 *pb.Fleet
	for _, f := range sourceSystem.Fleets {
		if f.Id == int32(fleetID) {
			fleet = f
			fleet2 = f
			break
		}
	}

	shipType := ShipType(parts[2])
	if !shipType.IsValid() {
		return CommandResult{
			Success: false,
			Message: "Invalid ship type: <de|cr|ba|dr>",
		}
	}
	log.Printf("Fleet: %v", fleet)
	log.Printf("Composition: %v", fleet.Composition)
	currentComposition := &pb.FleetComposition{
		Destroyers: fleet.Composition.Destroyers,
		Cruisers: fleet.Composition.Cruisers,
		Battleships: fleet.Composition.Battleships,
		Dreadnoughts: fleet.Composition.Dreadnoughts,
	}

	switch shipType {
	case Destroyer:
		currentComposition.Destroyers += 1
	case Cruiser:
		currentComposition.Cruisers += 1
	case Battleship:
		currentComposition.Battleships += 1
	case Dreadnought:
		currentComposition.Dreadnoughts += 1
	}
	log.Printf("Composition AFTER: %v", currentComposition)
	log.Printf("Composition Original: %v", fleet2.Composition)
	modification := &pb.FleetModification{
		FleetId: int32(fleetID),
		SystemId: systemID,
		Composition: currentComposition,
		Owner: client.username,
	}

	err = client.SendFleetModification(modification)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Failed to send fleet modification: %v", err),
		}
	}

	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("Fleet modification request sent for system %d", systemID),
	}
}

func handleFleetCreation(client *UDPClient, parts []string) CommandResult {
	if len(parts) != 2 {
		return CommandResult{
			Success: false,
			Message: "Invalid fleet creation command. Format: fc <system_id>",
		}
	}

	systemID, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Invalid system ID: %s", parts[1]),
		}
	}

	creation := &pb.FleetCreation{
		SystemId: int32(systemID),
		Owner:    client.username,
	}

	err = client.SendFleetCreation(creation)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Failed to send fleet creation: %v", err),
		}
	}

	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("Fleet created in system %d", systemID),
	}
}

// Format: fm <fleet_id> <to_system_id>
func handleFleetMovement(client *UDPClient, parts []string, data *CommandData) CommandResult {
	if len(parts) != 3 {
		return CommandResult{
			Success: false,
			Message: "Invalid fleet movement command. Format: fm <fleet_id> <to_system_id>",
		}
	}

	fleetID, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Invalid fleet ID: %s", parts[1]),
		}
	}

	toSystemID, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Invalid destination system ID: %s", parts[3]),
		}
	}

	fromSystemID := data.FleetLocations[int32(fleetID)]
	movement := &pb.FleetMovement{
		FleetId:      int32(fleetID),
		FromSystemId: int32(fromSystemID),
		ToSystemId:   int32(toSystemID),
	}

	err = client.SendFleetMovement(movement)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Failed to send fleet movement: %v", err),
		}
	}

	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("Fleet %d is moving from system %d to system %d", fleetID, fromSystemID, toSystemID),
	}
}
