package memcached

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const maxQueueLength = 20000

type TcpConfig struct {
	Addr        string
	MaxIdleConn int
	MaxOpenConn int
}

type TcpConnPool struct {
	addr         string
	mu           sync.Mutex
	idleConns    map[string]*connection
	numOpen      int
	maxOpenCount int
	maxIdleCount int
	requestChan  chan *connRequest
}

func NewConnPool(cfg *TcpConfig) (*TcpConnPool, error) {
	pool := &TcpConnPool{
		addr:         cfg.Addr,
		idleConns:    make(map[string]*connection),
		requestChan:  make(chan *connRequest, maxQueueLength),
		maxOpenCount: cfg.MaxOpenConn,
		maxIdleCount: cfg.MaxIdleConn,
	}

	go pool.handleConnectionRequest()

	return pool, nil
}

type connRequest struct {
	connChan chan *connection
	errChan  chan error
}

func (t *TcpConnPool) put(c *connection) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.maxIdleCount > 0 && t.maxIdleCount > len(t.idleConns) {
		t.idleConns[c.id] = c
	} else {
		c.nc.Close()
		c.pool.numOpen--
	}
}

func (t *TcpConnPool) get() (*connection, error) {
	t.mu.Lock()

	numIdle := len(t.idleConns)
	if numIdle > 0 {

		for _, c := range t.idleConns {

			delete(t.idleConns, c.id)
			t.mu.Unlock()
			return c, nil
		}
	}
	if t.maxOpenCount > 0 && t.numOpen >= t.maxOpenCount {

		req := &connRequest{
			connChan: make(chan *connection, 1),
			errChan:  make(chan error, 1),
		}

		t.requestChan <- req

		t.mu.Unlock()

		select {
		case tcpConn := <-req.connChan:
			return tcpConn, nil
		case err := <-req.errChan:
			return nil, err
		}
	}
	t.numOpen++
	t.mu.Unlock()

	newTcpConn, err := t.openNewTcpConnection()
	if err != nil {
		t.mu.Lock()
		t.numOpen--
		t.mu.Unlock()
		return nil, err
	}

	return newTcpConn, nil
}

func (t *TcpConnPool) openNewTcpConnection() (*connection, error) {

	nc, err := net.Dial("tcp", t.addr)
	if err != nil {
		return nil, err
	}

	return &connection{
		// Use unix time as id
		id:   fmt.Sprintf("%v", time.Now().UnixNano()),
		nc:   nc,
		pool: t,
		rw:   bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}, nil
}

func (t *TcpConnPool) handleConnectionRequest() {
	for req := range t.requestChan {
		var (
			requestDone = false
			hasTimeout  = false
			timeoutChan = time.After(3 * time.Second)
		)

		for {
			if requestDone || hasTimeout {
				break
			}
			select {
			// request timeout
			case <-timeoutChan:
				hasTimeout = true
				req.errChan <- errors.New("connection request timeout")
			default:
				t.mu.Lock()
				numIdle := len(t.idleConns)
				switch {
				case numIdle > 0:

					for _, c := range t.idleConns {
						delete(t.idleConns, c.id)
						t.mu.Unlock()
						req.connChan <- c // give conn
						requestDone = true
						break
					}
				case t.maxOpenCount > 0 && t.numOpen < t.maxOpenCount:
					t.numOpen++
					t.mu.Unlock()

					c, err := t.openNewTcpConnection()
					if err != nil {
						t.mu.Lock()
						t.numOpen--
						t.mu.Unlock()
					} else {
						req.connChan <- c // give conn
						requestDone = true
					}
				default:
					t.mu.Unlock()

				}
			}
		}
	}
}

func (t *TcpConnPool) Close() {
	if t.idleConns == nil {
		return
	}
	for _, conn := range t.idleConns {
		conn.Close()
	}
}
