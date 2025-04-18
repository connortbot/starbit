package game

import (
	"fmt"
	"strconv"
	"strings"

	pb "starbit/proto"
)

type CommandResult struct {
	Success bool
	Message string
}

// data needed for parsing all commands
type CommandData struct {
	FleetLocations map[int32]int32
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
	default:
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Unknown command: %s", cmdType),
		}
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
