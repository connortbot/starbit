package game

import (
	"fmt"
	"slices"
	"sync"

	pb "starbit/proto"

	galaxy "starbit/server/game/galaxy"
)

const (
	galaxyWidth  = 5
	galaxyHeight = 5

	initialGES = 1000
	gesPerTick = 2
	fleetCost  = 2000

	FLEET_MOVEMENT_COOLDOWN = 10
)

var MaxPlayers int32 = 4

func SetMaxPlayers(n int32) {
	if n > 0 {
		MaxPlayers = n
	}
}

type State struct {
	PlayerCount int32
	Players     map[string]*pb.Player
	Started     bool
	Galaxy      *pb.GalaxyState
	mu          sync.Mutex
	nextFleetID int32

	playerGES map[string]int32

	movedFleets     []int32 // fleets who have moved this tick
	battlingSystems []int32 // systems who are currently battling

	// Map of username -> array of system IDs they own
	ownedSystems map[string][]int32
	tickCount    int32
}

func NewState() *State {
	systems := make([]*pb.System, 0, galaxyWidth*galaxyHeight)
	id := 0
	for y := 0; y < galaxyHeight; y++ {
		for x := 0; x < galaxyWidth; x++ {
			systems = append(systems, &pb.System{
				Id:    int32(id),
				X:     int32(x),
				Y:     int32(y),
				Owner: "none",
			})
			id += 1
		}
	}

	return &State{
		PlayerCount: 0,
		Players:     make(map[string]*pb.Player),
		Started:     false,
		Galaxy: &pb.GalaxyState{
			Systems: systems,
			Width:   galaxyWidth,
			Height:  galaxyHeight,
		},
		nextFleetID:  1, // start fleet IDs at 1
		playerGES:    make(map[string]int32),
		ownedSystems: make(map[string][]int32),
		tickCount:    0,
	}
}

func (g *State) AddPlayer(name string) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if int32(len(g.Players)) >= MaxPlayers {
		return true, fmt.Errorf("game is full (max %d players)", MaxPlayers)
	}

	if _, exists := g.Players[name]; exists {
		return false, fmt.Errorf("username '%s' is already taken", name)
	}

	player := &pb.Player{
		Name: name,
	}
	g.Players[name] = player
	g.PlayerCount = int32(len(g.Players))

	g.playerGES[name] = initialGES

	if int32(len(g.Players)) == MaxPlayers {
		g.Started = true
		return true, nil
	}

	return false, nil
}

func (g *State) RemovePlayer(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Players, name)
	g.PlayerCount = int32(len(g.Players))
}

func (g *State) InitializeGame(players map[string]*pb.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.ownedSystems = make(map[string][]int32)
	for name := range players {
		g.ownedSystems[name] = []int32{}
	}

	galaxy.InitializeGalaxy(g.Galaxy, players, func() int32 {
		id := g.nextFleetID
		g.nextFleetID++
		return id
	})

	for _, system := range g.Galaxy.Systems {
		if system.Owner != "none" {
			g.ownedSystems[system.Owner] = append(g.ownedSystems[system.Owner], system.Id)
		}
	}

	g.Started = true
}

func (g *State) MoveFleet(username string, fleetMovement *pb.FleetMovement, tickCount int32) (*pb.SystemOwnerChange, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if slices.Contains(g.movedFleets, fleetMovement.FleetId) || slices.Contains(g.battlingSystems, fleetMovement.FromSystemId) {
		return nil, fmt.Errorf("fleet %d has already been moved or is battling", fleetMovement.FleetId)
	}

	if fleetMovement.FromSystemId < 0 || fleetMovement.FromSystemId >= int32(len(g.Galaxy.Systems)) {
		return nil, fmt.Errorf("source system with ID %d not found", fleetMovement.FromSystemId)
	}
	if fleetMovement.ToSystemId < 0 || fleetMovement.ToSystemId >= int32(len(g.Galaxy.Systems)) {
		return nil, fmt.Errorf("destination system with ID %d not found", fleetMovement.ToSystemId)
	}

	sourceSystem := g.Galaxy.Systems[fleetMovement.FromSystemId]
	destSystem := g.Galaxy.Systems[fleetMovement.ToSystemId]

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
		return nil, fmt.Errorf("fleet ID %d not found in system ID %d",
			fleetMovement.FleetId, fleetMovement.FromSystemId)
	}

	if fleet.LastMovedTick > (tickCount - FLEET_MOVEMENT_COOLDOWN) {
		ticksToWait := FLEET_MOVEMENT_COOLDOWN - (tickCount - fleet.LastMovedTick)
		return nil, fmt.Errorf("fleet %d has already moved this tick, wait %d ticks before moving again", fleetMovement.FleetId, ticksToWait)
	}

	if fleet.Owner != username {
		return nil, fmt.Errorf("fleet %d is not owned by %s", fleetMovement.FleetId, username)
	}

	// remove from source system
	sourceSystem.Fleets = append(sourceSystem.Fleets[:fleetIndex], sourceSystem.Fleets[fleetIndex+1:]...)
	// add to destination system
	destSystem.Fleets = append(destSystem.Fleets, fleet)
	fleet.LastMovedTick = tickCount
	g.movedFleets = append(g.movedFleets, fleetMovement.FleetId)

	if galaxy.ShouldBattleBegin(g.Galaxy, fleetMovement.ToSystemId) {
		if !slices.Contains(g.battlingSystems, fleetMovement.ToSystemId) {
			g.battlingSystems = append(g.battlingSystems, fleetMovement.ToSystemId)
		}
		return g.SetSystemOwner(fleetMovement.ToSystemId, "none")
	}

	// check only one player's fleets here
	ownerMap := make(map[string]bool)
	for _, f := range destSystem.Fleets {
		if f.Owner != "" {
			ownerMap[f.Owner] = true
		}
	}

	// if there's exactly one owner, set the system owner to that fleet owner
	var ownerChange *pb.SystemOwnerChange
	if len(ownerMap) == 1 {
		var owner string
		for o := range ownerMap {
			owner = o
			break
		}

		var err error
		ownerChange, err = g.SetSystemOwner(fleetMovement.ToSystemId, owner)
		if err != nil {
			fmt.Printf("Failed to set system owner: %v\n", err)
		}
	}

	return ownerChange, nil
}

func (g *State) SetSystemOwner(systemId int32, owner string) (*pb.SystemOwnerChange, error) {
	if systemId < 0 || systemId >= int32(len(g.Galaxy.Systems)) {
		return nil, fmt.Errorf("system with ID %d not found", systemId)
	}

	currentOwner := g.Galaxy.Systems[systemId].Owner
	if currentOwner != owner {
		g.Galaxy.Systems[systemId].Owner = owner

		// update owned systems tracking
		if currentOwner != "none" {
			// remove system from previous owner's list
			for i, sys := range g.ownedSystems[currentOwner] {
				if sys == systemId {
					g.ownedSystems[currentOwner] = append(g.ownedSystems[currentOwner][:i], g.ownedSystems[currentOwner][i+1:]...)
					break
				}
			}
		}

		// add system to new owner's list (if not "none")
		if owner != "none" {
			if _, exists := g.ownedSystems[owner]; !exists {
				g.ownedSystems[owner] = []int32{}
			}
			g.ownedSystems[owner] = append(g.ownedSystems[owner], systemId)
		}

		return &pb.SystemOwnerChange{
			SystemId: systemId,
			Owner:    owner,
		}, nil
	}
	return nil, nil
}

func (g *State) GetPlayerGES(player string) int32 {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.playerGES[player]
}

func (g *State) AdjustPlayerGES(player string, amount int32) int32 {
	if _, exists := g.playerGES[player]; exists {
		g.playerGES[player] += amount
		return g.playerGES[player]
	}

	return 0
}
