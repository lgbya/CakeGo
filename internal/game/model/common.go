package model

import "cake/proto/pb"

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p *Pos) Pb() *pb.Pos {
	return &pb.Pos{X: uint32(p.X), Y: uint32(p.Y)}
}
