package tcp

import (
	"bufio"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"net"
	"time"
)

type TcpConn struct {
	conn     net.Conn
	writeBuf *bufio.Writer
}

func NewTcpConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		conn:     conn,
		writeBuf: bufio.NewWriterSize(conn, 4096),
	}
}

func (c *TcpConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *TcpConn) Send(buf []byte) {
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

func (c *TcpConn) Close(id uint32) error {
	err := c.conn.Close()
	if err != nil {
		logger.Errorf("[%d]关闭网络错误：%v", id, err)
		return err
	}
	return nil
}
