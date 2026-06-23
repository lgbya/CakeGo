package testsvc

import (
	"cake/internal/gensvc/rpc"
	"fmt"
)

type State struct {
}

type Service struct {
	*rpc.Service
}

func StartService() (*rpc.Service, error) {
	s := &Service{}
	cfg := rpc.NewCfg()
	cfg.SendMaxCap = 1
	roleRpc, err := rpc.StartWithCfg("test", s, cfg)
	return roleRpc, err
}

func (s *Service) SvcName() string {
	return "test"
}

func (s *Service) Init(r *rpc.Service, args any) (any, error) {
	s.Service = r
	return &State{}, nil
}

func (s *Service) RpcTest(state any, args any) (any, error) {
	fmt.Println("send_test", args)
	return nil, nil
}

func (s *Service) Stop(_ any) {

}

var _ rpc.GenService = &Service{}
