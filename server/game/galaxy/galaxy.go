package galaxy

import (
	"math/rand"
	"slices"
	pb "starbit/proto"
	fleets "starbit/server/game/fleets"
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

	basicComposition := &pb.FleetComposition{
		Destroyers: 1,
	}

	p1Fleet := fleets.NewFleet(generateFleetID(), playerNames[0], fleets.StartingFleetAttack, fleets.StartingFleetExAttack, fleets.StartingFleetHealth, fleets.StartingFleetEvasion, fleets.StartingFleetArmor, basicComposition)
	p2Fleet := fleets.NewFleet(generateFleetID(), playerNames[1], fleets.StartingFleetAttack, fleets.StartingFleetExAttack, fleets.StartingFleetHealth, fleets.StartingFleetEvasion, fleets.StartingFleetArmor, basicComposition)

	if len(players) == 2 {
		SetSystemOwner(galaxy, topLeftCornerIndex, playerNames[0])
		AddFleetToSystem(galaxy, topLeftCornerIndex, p1Fleet)
		SetSystemOwner(galaxy, bottomRightCornerIndex, playerNames[1])
		AddFleetToSystem(galaxy, bottomRightCornerIndex, p2Fleet)
	} else if len(players) == 3 {
		p3Fleet := fleets.NewFleet(generateFleetID(), playerNames[2], fleets.StartingFleetAttack, fleets.StartingFleetExAttack, fleets.StartingFleetHealth, fleets.StartingFleetEvasion, fleets.StartingFleetArmor, basicComposition)
		SetSystemOwner(galaxy, topLeftCornerIndex, playerNames[0])
		AddFleetToSystem(galaxy, topLeftCornerIndex, p1Fleet)
		SetSystemOwner(galaxy, topRightCornerIndex, playerNames[1])
		AddFleetToSystem(galaxy, topRightCornerIndex, p2Fleet)
		SetSystemOwner(galaxy, bottomLeftCornerIndex, playerNames[2])
		AddFleetToSystem(galaxy, bottomLeftCornerIndex, p3Fleet)
	} else if len(players) == 4 {
		p3Fleet := fleets.NewFleet(generateFleetID(), playerNames[2], fleets.StartingFleetAttack, fleets.StartingFleetExAttack, fleets.StartingFleetHealth, fleets.StartingFleetEvasion, fleets.StartingFleetArmor, basicComposition)
		p4Fleet := fleets.NewFleet(generateFleetID(), playerNames[3], fleets.StartingFleetAttack, fleets.StartingFleetExAttack, fleets.StartingFleetHealth, fleets.StartingFleetEvasion, fleets.StartingFleetArmor, basicComposition)
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

func SetSystemOwner(galaxy *pb.GalaxyState, id int32, owner string) *pb.SystemOwnerChange {
	if galaxy.Systems[id].Owner != owner {
		galaxy.Systems[id].Owner = owner
		return &pb.SystemOwnerChange{
			SystemId: id,
			Owner:    owner,
		}
	}
	return nil
}

func AddFleetToSystem(galaxy *pb.GalaxyState, id int32, fleet *pb.Fleet) {
	galaxy.Systems[id].Fleets = append(galaxy.Systems[id].Fleets, fleet)
}

// returns true if a battle should begin between the fleets in the given system
func ShouldBattleBegin(galaxy *pb.GalaxyState, systemId int32) bool {
	system := galaxy.Systems[systemId]
	ownerMap := make(map[string]bool)

	for _, fleet := range system.Fleets {
		if fleet.Owner != "" && fleet.Health > 0 {
			ownerMap[fleet.Owner] = true
		}
	}
	return len(ownerMap) >= 2
}

func ExecuteBattle(galaxy *pb.GalaxyState, systemId int32) (bool, []*pb.HealthUpdate, []*pb.FleetDestroyed, string) {
	system := galaxy.Systems[systemId]
	updates := []*pb.HealthUpdate{}
	destroyed := []*pb.FleetDestroyed{}
	updatedFleets := []*pb.Fleet{}

	battleActive, newOwner := DetermineSystemOwner(galaxy, systemId)
	if !battleActive {
		return battleActive, updates, destroyed, newOwner
	}

	// group fleets by owner
	simpleFleetsByOwner := make(map[string][]*pb.Fleet)
	for _, fleet := range system.Fleets {
		if fleet.Health > 0 {
			simpleFleetsByOwner[fleet.Owner] = append(simpleFleetsByOwner[fleet.Owner], fleet)
		}
	}

	type FleetGroups struct {
		Own   []*pb.Fleet
		Enemy []*pb.Fleet
	}
	fleetsByOwner := make(map[string]*FleetGroups)

	for owner := range simpleFleetsByOwner {
		fleetsByOwner[owner] = &FleetGroups{
			Own:   simpleFleetsByOwner[owner],
			Enemy: make([]*pb.Fleet, 0),
		}
	}

	for owner, groups := range fleetsByOwner {
		for enemyOwner, enemyFleets := range simpleFleetsByOwner {
			if enemyOwner != owner {
				groups.Enemy = append(groups.Enemy, enemyFleets...)
			}
		}
	}

	for _, groups := range fleetsByOwner {
		for _, fleet := range groups.Own {
			if len(groups.Enemy) == 0 {
				continue // skip if no enemies to attack
			}
			enemyIndex := randomInt32(0, int32(len(groups.Enemy)))
			enemyFleet := groups.Enemy[enemyIndex]

			// Phase A: Apply Attack
			effectiveAtk := fleet.Attack * (1 - (enemyFleet.Armor / 100))
			if randomInt32(0, 100) > enemyFleet.Evasion {
				enemyFleet.Health -= effectiveAtk
			}
			
			// Phase B: Apply ExAttack
			effectiveExatk := fleet.Exattack * (1 - (enemyFleet.Armor / 100))
			enemyFleet.Health -= effectiveExatk

			if !slices.Contains(updatedFleets, enemyFleet) {
				updatedFleets = append(updatedFleets, enemyFleet)
			}
		}
	}

	fleetsToRemove := []*pb.Fleet{}
	for _, fleet := range updatedFleets {
		if fleet.Health <= 0 {
			destroyed = append(destroyed, &pb.FleetDestroyed{
				FleetId:  fleet.Id,
				SystemId: systemId,
			})
			fleetsToRemove = append(fleetsToRemove, fleet)
		} else {
			updates = append(updates, &pb.HealthUpdate{
				FleetId:  fleet.Id,
				Health:   fleet.Health,
				SystemId: systemId,
			})
		}
	}

	// Remove destroyed fleets from the system
	if len(fleetsToRemove) > 0 {
		newFleets := make([]*pb.Fleet, 0, len(system.Fleets)-len(fleetsToRemove))
		for _, fleet := range system.Fleets {
			shouldRemove := false
			for _, deadFleet := range fleetsToRemove {
				if fleet.Id == deadFleet.Id {
					shouldRemove = true
					break
				}
			}
			if !shouldRemove {
				newFleets = append(newFleets, fleet)
			}
		}
		system.Fleets = newFleets
	}

	battleActive, newOwner = DetermineSystemOwner(galaxy, systemId)
	return battleActive, updates, destroyed, newOwner
}

func DetermineSystemOwner(galaxy *pb.GalaxyState, systemId int32) (bool, string) {
	system := galaxy.Systems[systemId]
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

	return battleActive, newOwner
}

func randomInt32(min, max int32) int32 {
	if min == max {
		return min
	}
	// for array indices, we want to generate a number in range [min, max)
	// where the upper bound is exclusive
	return min + rand.Int31n(max-min)
}
