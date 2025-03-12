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

func ApplyHealthUpdate(galaxy *pb.GalaxyState, healthUpdate *pb.HealthUpdate) error {
	if healthUpdate.SystemId < 0 || healthUpdate.SystemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("system with ID %d not found", healthUpdate.SystemId)
	}

	system := galaxy.Systems[healthUpdate.SystemId]
	for _, fleet := range system.Fleets {
		if fleet.Id == healthUpdate.FleetId {
			fleet.Health = healthUpdate.Health
			return nil
		}
	}

	return fmt.Errorf("fleet ID %d not found in system ID %d", healthUpdate.FleetId, healthUpdate.SystemId)
}

func ProcessFleetDestroyed(galaxy *pb.GalaxyState, fleetDestroyed *pb.FleetDestroyed) error {
	if fleetDestroyed.SystemId < 0 || fleetDestroyed.SystemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("system with ID %d not found", fleetDestroyed.SystemId)
	}

	system := galaxy.Systems[fleetDestroyed.SystemId]
	for i, fleet := range system.Fleets {
		if fleet.Id == fleetDestroyed.FleetId {
			system.Fleets = append(system.Fleets[:i], system.Fleets[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("fleet ID %d not found in system ID %d", fleetDestroyed.FleetId, fleetDestroyed.SystemId)
}

func NewFleet(fleetId int32, owner string, attack int32, health int32) *pb.Fleet {
	return &pb.Fleet{
		Id:     fleetId,
		Owner:  owner,
		Attack: attack,
		Health: health,
	}
}

func AddFleetToSystem(galaxy *pb.GalaxyState, systemId int32, fleet *pb.Fleet) error {
	if systemId < 0 || systemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("system with ID %d not found", systemId)
	}
	galaxy.Systems[systemId].Fleets = append(galaxy.Systems[systemId].Fleets, fleet)
	return nil
}


func ProcessFleetCreation(galaxy *pb.GalaxyState, fleetCreation *pb.FleetCreation) error {
	if fleetCreation.SystemId < 0 || fleetCreation.SystemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("system with ID %d not found", fleetCreation.SystemId)
	}

	AddFleetToSystem(
		galaxy,
		fleetCreation.SystemId,
		NewFleet(
			fleetCreation.FleetId,
			fleetCreation.Owner,
			fleetCreation.Attack,
			fleetCreation.Health,
		),
	)
	return nil
}

func SetSystemOwner(galaxy *pb.GalaxyState, systemId int32, owner string) error {
	if systemId < 0 || systemId >= int32(len(galaxy.Systems)) {
		return fmt.Errorf("system with ID %d not found", systemId)
	}
	galaxy.Systems[systemId].Owner = owner
	return nil
}
