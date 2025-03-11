package game

import (
	"fmt"
	"sync"

	pb "starbit/proto"
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

	g.initializeGalaxy(players)
	g.Started = true
}

func (g *State) initializeGalaxy(players map[string]*pb.Player) {
	playerNames := make([]string, 0, len(players))
	for name := range players {
		playerNames = append(playerNames, name)
	}

	topLeftCornerIndex := int32(0)
	topRightCornerIndex := g.Galaxy.Width - 1
	bottomLeftCornerIndex := g.Galaxy.Width * (g.Galaxy.Height - 1)
	bottomRightCornerIndex := g.Galaxy.Width * g.Galaxy.Height - 1

	if len(players) == 2 {
		g.setSystemOwner(topLeftCornerIndex, playerNames[0])
		g.setSystemOwner(bottomRightCornerIndex, playerNames[1])
	} else if len(players) == 3 {
		g.setSystemOwner(topLeftCornerIndex, playerNames[0])
		g.setSystemOwner(topRightCornerIndex, playerNames[1])
		g.setSystemOwner(bottomLeftCornerIndex, playerNames[2])
	} else if len(players) == 4 {
		g.setSystemOwner(topLeftCornerIndex, playerNames[0])
		g.setSystemOwner(topRightCornerIndex, playerNames[1])
		g.setSystemOwner(bottomLeftCornerIndex, playerNames[2])
		g.setSystemOwner(bottomRightCornerIndex, playerNames[3])
	}
}

func (g *State) setSystemOwner(id int32, owner string) {
	g.Galaxy.Systems[id].Owner = owner
}
