package sdk

import (
	"cake/internal/game/def/errcode"
	"cake/internal/util/errx"
	"cake/proto/pb"
)

var authChannel = make(map[uint32]authenticator)

type authenticator interface {
	Auth(*pb.AccountAuthC2S) (err error)
}

func Init() {
	RegisterAuth(ChannelIdTemp, new(Template))
}

func RegisterAuth(channelID uint32, auth authenticator) {
	authChannel[channelID] = auth
}

func AuthChannel(reqData *pb.AccountAuthC2S) error {
	channelID := reqData.ChannelId
	auth, ok := authChannel[channelID]
	if !ok {
		return errx.New(errcode.LoginSdkFail, "", "")
	}

	if err := auth.Auth(reqData); err != nil {
		return errx.From(err)
	}
	return nil
}
