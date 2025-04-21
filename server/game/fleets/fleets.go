package fleets

import (
	pb "starbit/proto"
)

const (
	// Destroyer stats
	DestroyerCost     = 250
	DestroyerHealth   = 50
	DestroyerAttack   = 2
	DestroyerExAttack = 1
	DestroyerEvasion  = 35
	DestroyerArmor    = 5

	// Cruiser stats
	CruiserCost     = 350
	CruiserHealth   = 75
	CruiserAttack   = 1
	CruiserExAttack = 2
	CruiserEvasion  = 20
	CruiserArmor    = 15

	// Battleship stats
	BattleshipCost     = 800
	BattleshipHealth   = 200
	BattleshipAttack   = 5
	BattleshipExAttack = 2
	BattleshipEvasion  = 10
	BattleshipArmor    = 30

	// Dreadnought stats
	DreadnoughtCost     = 1500
	DreadnoughtHealth   = 600
	DreadnoughtAttack   = 3
	DreadnoughtExAttack = 5
	DreadnoughtEvasion  = 5
	DreadnoughtArmor    = 40

	StartingFleetHealth = DestroyerHealth
	StartingFleetAttack = DestroyerAttack
	StartingFleetExAttack = DestroyerExAttack
	StartingFleetEvasion = DestroyerEvasion
	StartingFleetArmor = DestroyerArmor
)

func NewFleet(fleetId int32, owner string, attack int32, exattack int32, health int32, evasion int32, armor int32, composition *pb.FleetComposition) *pb.Fleet {
	return &pb.Fleet{
		Id:     fleetId,
		Owner:  owner,
		Attack: attack,
		Health: health,
		MaxHealth: health,
		Exattack: exattack,
		Evasion: evasion,
		Armor: armor,
		Composition: composition,
	}
}

// recalculates the stats and applies them and assumes that the composition has not already been applied
func CalculateAndUpdateFleet(fleet *pb.Fleet, fleetModification *pb.FleetModification) {
	modificationHealth := int32(0)
	modificationHealth += int32(DestroyerHealth) * (fleetModification.Composition.Destroyers - fleet.Composition.Destroyers)	
	modificationHealth += int32(CruiserHealth) * (fleetModification.Composition.Cruisers - fleet.Composition.Cruisers)
	modificationHealth += int32(BattleshipHealth) * (fleetModification.Composition.Battleships - fleet.Composition.Battleships)
	modificationHealth += int32(DreadnoughtHealth) * (fleetModification.Composition.Dreadnoughts - fleet.Composition.Dreadnoughts)
	
	
	fleet.Health += modificationHealth
	fleet.Composition = fleetModification.Composition

	fleet.MaxHealth = int32(DestroyerHealth) * fleet.Composition.Destroyers + 
		int32(CruiserHealth) * fleet.Composition.Cruisers + 
		int32(BattleshipHealth) * fleet.Composition.Battleships + 
		int32(DreadnoughtHealth) * fleet.Composition.Dreadnoughts	
	fleet.Attack = int32(DestroyerAttack) * fleet.Composition.Destroyers + 
		int32(CruiserAttack) * fleet.Composition.Cruisers + 
		int32(BattleshipAttack) * fleet.Composition.Battleships + 
		int32(DreadnoughtAttack) * fleet.Composition.Dreadnoughts
	fleet.Exattack = int32(DestroyerExAttack) * fleet.Composition.Destroyers + 
		int32(CruiserExAttack) * fleet.Composition.Cruisers + 
		int32(BattleshipExAttack) * fleet.Composition.Battleships + 
		int32(DreadnoughtExAttack) * fleet.Composition.Dreadnoughts

	totalShips := fleet.Composition.Destroyers + fleet.Composition.Cruisers + fleet.Composition.Battleships + fleet.Composition.Dreadnoughts
	
	if totalShips == 0 {
		fleet.Evasion = 0
		fleet.Armor = 0
		return
	}

	totalEvasion := int32(DestroyerEvasion) * fleet.Composition.Destroyers + 
		int32(CruiserEvasion) * fleet.Composition.Cruisers + 
		int32(BattleshipEvasion) * fleet.Composition.Battleships + 
		int32(DreadnoughtEvasion) * fleet.Composition.Dreadnoughts

	totalArmor := int32(DestroyerArmor) * fleet.Composition.Destroyers + 
		int32(CruiserArmor) * fleet.Composition.Cruisers + 
		int32(BattleshipArmor) * fleet.Composition.Battleships + 
		int32(DreadnoughtArmor) * fleet.Composition.Dreadnoughts

	fleet.Evasion = (totalEvasion + totalShips/2) / totalShips
	fleet.Armor = (totalArmor + totalShips/2) / totalShips
}


