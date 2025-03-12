package game

import (
	"fmt"
	"sync"

	pb "starbit/proto"

	galaxy "starbit/server/game/galaxy"
)

const (
	galaxyWidth  = 5
	galaxyHeight = 5
	maxPlayers   = 2
)

type State struct {
	PlayerCount int32
	Players     map[string]*pb.Player
	Started     bool
	Galaxy      *pb.GalaxyState
	mu          sync.Mutex
	nextFleetID int32
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

func (g *State) MoveFleet(username string, fleetMovement *pb.FleetMovement) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if fleetMovement.FromSystemId < 0 || fleetMovement.FromSystemId >= int32(len(g.Galaxy.Systems)) {
		return fmt.Errorf("source system with ID %d not found", fleetMovement.FromSystemId)
	}
	if fleetMovement.ToSystemId < 0 || fleetMovement.ToSystemId >= int32(len(g.Galaxy.Systems)) {
		return fmt.Errorf("destination system with ID %d not found", fleetMovement.ToSystemId)
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
		return fmt.Errorf("fleet ID %d not found in system ID %d",
			fleetMovement.FleetId, fleetMovement.FromSystemId)
	}

	if fleet.Owner != username {
		return fmt.Errorf("fleet %d is not owned by %s", fleetMovement.FleetId, username)
	}

	// remove from source system
	sourceSystem.Fleets = append(sourceSystem.Fleets[:fleetIndex], sourceSystem.Fleets[fleetIndex+1:]...)
	// add to destination system
	destSystem.Fleets = append(destSystem.Fleets, fleet)
	return nil
}
