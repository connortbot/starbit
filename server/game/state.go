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
	maxPlayers   = 2

	initialGES = 1000
	gesPerTick = 2
)

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
		nextFleetID: 1, // start fleet IDs at 1
		playerGES:   make(map[string]int32),
	}
}

func (g *State) AddPlayer(name string) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.Players) >= maxPlayers {
		return true, fmt.Errorf("game is full (max %d players)", maxPlayers)
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

	if len(g.Players) == maxPlayers {
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

	galaxy.InitializeGalaxy(g.Galaxy, players, func() int32 {
		id := g.nextFleetID
		g.nextFleetID++
		return id
	})
	g.Started = true
}

func (g *State) MoveFleet(username string, fleetMovement *pb.FleetMovement) (*pb.SystemOwnerChange, error) {
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

	if fleet.Owner != username {
		return nil, fmt.Errorf("fleet %d is not owned by %s", fleetMovement.FleetId, username)
	}

	// remove from source system
	sourceSystem.Fleets = append(sourceSystem.Fleets[:fleetIndex], sourceSystem.Fleets[fleetIndex+1:]...)
	// add to destination system
	destSystem.Fleets = append(destSystem.Fleets, fleet)
	g.movedFleets = append(g.movedFleets, fleetMovement.FleetId)

	if galaxy.ShouldBattleBegin(g.Galaxy, fleetMovement.ToSystemId) {
		g.battlingSystems = append(g.battlingSystems, fleetMovement.ToSystemId)
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

	if g.Galaxy.Systems[systemId].Owner != owner {
		g.Galaxy.Systems[systemId].Owner = owner

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
