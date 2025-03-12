package game

import (
	"fmt"
	pb "starbit/proto"
)

func MoveFleet(galaxy *pb.GalaxyState, fleetMovement *pb.FleetMovement) error {
	if fleetMovement.FromSystemId < 0 || fleetMovement.FromSystemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("source system with ID %d not found", fleetMovement.FromSystemId)
	}
	if fleetMovement.ToSystemId < 0 || fleetMovement.ToSystemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("destination system with ID %d not found", fleetMovement.ToSystemId)
	}

	sourceSystem := galaxy.Systems[fleetMovement.FromSystemId]
	destSystem := galaxy.Systems[fleetMovement.ToSystemId]

	var fleet *pb.Fleet
	var fleetIndex int
	for i, f := range sourceSystem.Fleets {
		if f.Id == fleetMovement.FleetId {
			fleet = f
			fleetIndex = i
			break
		}
	}

	if fleet == nil {
		return fmt.Errorf("fleet ID %d not found in system ID %d",
			fleetMovement.FleetId, fleetMovement.FromSystemId)
	}

	sourceSystem.Fleets = append(sourceSystem.Fleets[:fleetIndex], sourceSystem.Fleets[fleetIndex+1:]...)
	destSystem.Fleets = append(destSystem.Fleets, fleet)

	return nil
}
