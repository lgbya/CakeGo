package connsvc

type IConn interface {
	Close(id uint32) error
	Read() ([]byte, error)
	Send([]byte)
	Heartbeat()
}
