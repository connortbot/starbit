package fleets

import (
	pb "starbit/proto"
)

func NewFleet(owner string, attack int32, health int32) *pb.Fleet {
	return &pb.Fleet{
		Owner:  owner,
		Attack: attack,
		Health: health,
	}
}
