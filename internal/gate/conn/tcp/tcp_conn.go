package tcp

import (
	"bufio"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"net"
	"time"
)

type Conn struct {
	conn     net.Conn
	writeBuf *bufio.Writer
	buf      []byte
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		buf:      make([]byte, 1024),
		conn:     conn,
		writeBuf: bufio.NewWriterSize(conn, 4096),
	}
}

func (c *Conn) Read() ([]byte, error) {
	n, err := c.conn.Read(c.buf)
	if err != nil {
		return nil, err
	}
	// 复制数据，防止 buf 被复用污染
	data := make([]byte, n)
	copy(data, c.buf[:n])
	return data, nil
}

func (c *Conn) Send(buf []byte) {
	defer sys.Recover("SendPacket")
	err := c.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		logger.Errorf("set write dead line err %v", err)
		return
	}

	_, err = c.writeBuf.Write(buf)
	if err != nil {
		logger.Errorf("write data err %v", err)
		return
	}
	// 强制刷到网络，小包立刻发送
	if err = c.writeBuf.Flush(); err != nil {
		logger.Errorf("flush data err %v", err)
		return
	}
}

func (c *Conn) Close(id uint32) error {
	err := c.conn.Close()
	if err != nil {
		logger.Errorf("[%d]关闭网络错误：%v", id, err)
		return err
	}
	return nil
}
