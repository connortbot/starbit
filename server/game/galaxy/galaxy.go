package galaxy

import (
	pb "starbit/proto"
	fleets "starbit/server/game/fleets"
)

const (
	StartingFleetHealth = 100
	StartingFleetAttack = 10
)

func InitializeGalaxy(galaxy *pb.GalaxyState, players map[string]*pb.Player, generateFleetID func() int32) {
	playerNames := make([]string, 0, len(players))
	for name := range players {
		playerNames = append(playerNames, name)
	}

	topLeftCornerIndex := int32(0)
	topRightCornerIndex := galaxy.Width - 1
	bottomLeftCornerIndex := galaxy.Width * (galaxy.Height - 1)
	bottomRightCornerIndex := galaxy.Width*galaxy.Height - 1

	p1Fleet := fleets.NewFleet(generateFleetID(), playerNames[0], StartingFleetAttack, StartingFleetHealth)
	p2Fleet := fleets.NewFleet(generateFleetID(), playerNames[1], StartingFleetAttack, StartingFleetHealth)

	if len(players) == 2 {
		SetSystemOwner(galaxy, topLeftCornerIndex, playerNames[0])
		AddFleetToSystem(galaxy, topLeftCornerIndex, p1Fleet)
		SetSystemOwner(galaxy, bottomRightCornerIndex, playerNames[1])
		AddFleetToSystem(galaxy, bottomRightCornerIndex, p2Fleet)
	} else if len(players) == 3 {
		p3Fleet := fleets.NewFleet(generateFleetID(), playerNames[2], StartingFleetAttack, StartingFleetHealth)
		SetSystemOwner(galaxy, topLeftCornerIndex, playerNames[0])
		AddFleetToSystem(galaxy, topLeftCornerIndex, p1Fleet)
		SetSystemOwner(galaxy, topRightCornerIndex, playerNames[1])
		AddFleetToSystem(galaxy, topRightCornerIndex, p2Fleet)
		SetSystemOwner(galaxy, bottomLeftCornerIndex, playerNames[2])
		AddFleetToSystem(galaxy, bottomLeftCornerIndex, p3Fleet)
	} else if len(players) == 4 {
		p3Fleet := fleets.NewFleet(generateFleetID(), playerNames[2], StartingFleetAttack, StartingFleetHealth)
		p4Fleet := fleets.NewFleet(generateFleetID(), playerNames[3], StartingFleetAttack, StartingFleetHealth)
		SetSystemOwner(galaxy, topLeftCornerIndex, playerNames[0])
		AddFleetToSystem(galaxy, topLeftCornerIndex, p1Fleet)
		SetSystemOwner(galaxy, topRightCornerIndex, playerNames[1])
		AddFleetToSystem(galaxy, topRightCornerIndex, p2Fleet)
		SetSystemOwner(galaxy, bottomLeftCornerIndex, playerNames[2])
		AddFleetToSystem(galaxy, bottomLeftCornerIndex, p3Fleet)
		SetSystemOwner(galaxy, bottomRightCornerIndex, playerNames[3])
		AddFleetToSystem(galaxy, bottomRightCornerIndex, p4Fleet)
	}
}

func SetSystemOwner(galaxy *pb.GalaxyState, id int32, owner string) {
	galaxy.Systems[id].Owner = owner
}

func AddFleetToSystem(galaxy *pb.GalaxyState, id int32, fleet *pb.Fleet) {
	galaxy.Systems[id].Fleets = append(galaxy.Systems[id].Fleets, fleet)
}
