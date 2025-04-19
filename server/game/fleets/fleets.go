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
