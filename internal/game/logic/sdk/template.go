package sdk

import "cake/proto/pb"

type Template struct {
}

func (t *Template) Auth(_ *pb.AccountAuthC2S) error {
	return nil
}
