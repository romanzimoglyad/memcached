package memcached

import (
	"bufio"
	"net"
)

type connection struct {
	id   string
	nc   net.Conn
	rw   *bufio.ReadWriter
	pool *TcpConnPool
}

func (c *connection) Close() error {
	return c.nc.Close()
}
