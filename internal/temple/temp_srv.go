package temple

import (
	"cake/internal/gensvc/rpc"
)

type Service struct {
	*rpc.Service
}

type State struct{}

func (t *Service) Init(s *rpc.Service, args any) (any, error) {
	t.Service = s

	return &State{}, nil
}

func (t *Service) RpcTest(state State, args any) (any, int) {
	//fmt.Println("send_test", args, state)
	return "b", 0
}

func (t *Service) Stop(_ any) {

}

var _ rpc.GenService = &Service{}
