package fleets

import (
	pb "starbit/proto"
)

func NewFleet(id int32, owner string, attack int32, health int32) *pb.Fleet {
	return &pb.Fleet{
		Id:            id,
		Owner:         owner,
		Attack:        attack,
		Health:        health,
		LastMovedTick: 0,
	}
}
