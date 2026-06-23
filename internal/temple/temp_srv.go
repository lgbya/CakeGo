package temple

import (
	"cake/internal/gensvc/rpc"
)

type Service struct {
	*rpc.Service
}

type State struct{}

func (s *Service) Init(rpcSvc *rpc.Service, args any) (any, error) {
	s.Service = rpcSvc

	return &State{}, nil
}

func (s *Service) RpcTest(state State, args any) (any, int) {
	//fmt.Println("send_test", args, state)
	return "b", 0
}

func (s *Service) Stop(_ any) {

}

var _ rpc.GenService = &Service{}
