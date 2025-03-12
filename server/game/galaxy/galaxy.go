package galaxy

import (
	"math/rand"
	pb "starbit/proto"
	fleets "starbit/server/game/fleets"
)

const (
	StartingFleetHealth = 100
	StartingFleetAttack = 5
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

// returns true if a battle should begin between the fleets in the given system
func ShouldBattleBegin(galaxy *pb.GalaxyState, systemId int32) bool {
	system := galaxy.Systems[systemId]
	ownerMap := make(map[string]bool)

	for _, fleet := range system.Fleets {
		if fleet.Owner != "" {
			ownerMap[fleet.Owner] = true
		}
	}
	return len(ownerMap) >= 2
}

func ExecuteBattle(galaxy *pb.GalaxyState, systemId int32) (bool, []*pb.HealthUpdate, []*pb.FleetDestroyed, string) {
	system := galaxy.Systems[systemId]
	updates := []*pb.HealthUpdate{}
	destroyed := []*pb.FleetDestroyed{}

	// group fleets by owner
	fleetsByOwner := make(map[string][]*pb.Fleet)
	for _, fleet := range system.Fleets {
		if fleet.Health > 0 {
			fleetsByOwner[fleet.Owner] = append(fleetsByOwner[fleet.Owner], fleet)
		}
	}

	// each fleet attacks a random enemy fleet
	for owner, fleets := range fleetsByOwner {
		for _, attackingFleet := range fleets {
			// create list of potential targets (fleets from other owners)
			potentialTargets := []*pb.Fleet{}
			for targetOwner, targetFleets := range fleetsByOwner {
				if targetOwner != owner {
					potentialTargets = append(potentialTargets, targetFleets...)
				}
			}

			if len(potentialTargets) > 0 {
				// choose a random target
				targetIndex := int(randomInt32(0, int32(len(potentialTargets))))
				targetFleet := potentialTargets[targetIndex]

				targetFleet.Health -= attackingFleet.Attack
				if targetFleet.Health <= 0 {
					targetFleet.Health = 0
					destroyed = append(destroyed, &pb.FleetDestroyed{
						FleetId:  targetFleet.Id,
						SystemId: systemId,
					})
					for i, fleet := range system.Fleets {
						if fleet.Id == targetFleet.Id {
							system.Fleets = append(system.Fleets[:i], system.Fleets[i+1:]...)
							break
						}
					}
				}

				updates = append(updates, &pb.HealthUpdate{
					FleetId:  targetFleet.Id,
					Health:   targetFleet.Health,
					SystemId: systemId,
				})
			}
		}
	}

	battleActive := ShouldBattleBegin(galaxy, systemId)

	newOwner := ""
	if !battleActive {
		ownerMap := make(map[string]bool)
		for _, fleet := range system.Fleets {
			if fleet.Health > 0 {
				ownerMap[fleet.Owner] = true
			}
		}

		if len(ownerMap) == 1 {
			for owner := range ownerMap {
				newOwner = owner
				break
			}
		} else if len(ownerMap) == 0 {
			newOwner = "none"
		}
	}

	return battleActive, updates, destroyed, newOwner
}

func randomInt32(min, max int32) int32 {
	if min == max {
		return min
	}
	return min + int32(float64(max-min)*rand.Float64())
}
