package game

import (
	"fmt"
	"sync"

	pb "starbit/proto"

	galaxy "starbit/server/game/galaxy"
)

const (
	galaxyWidth  = 10
	galaxyHeight = 10
	maxPlayers   = 2
)

type State struct {
	PlayerCount int32
	Players     map[string]*pb.Player
	Started     bool
	Galaxy      *pb.GalaxyState
	mu          sync.Mutex
}

func NewState() *State {
	systems := make([]*pb.System, 0, galaxyWidth*galaxyHeight)
	id := 0
	for y := 0; y < galaxyHeight; y++ {
		for x := 0; x < galaxyWidth; x++ {
			systems = append(systems, &pb.System{
				Id:  int32(id),
				X:   int32(x),
				Y:   int32(y),
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

	galaxy.InitializeGalaxy(g.Galaxy, players)
	g.Started = true
}