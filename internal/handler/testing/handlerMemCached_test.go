package testing

import (
	"bufio"
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/romanzimoglyad/memcached/internal/config"
	"github.com/romanzimoglyad/memcached/internal/handler"
	mem_cached "github.com/romanzimoglyad/memcached/internal/repository/mem-cached"
	"github.com/romanzimoglyad/memcached/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

type HandlerMemCachedSuite struct {
	suite.Suite
	RecordHandler *handler.RecordHandler
	cfg           *config.Config
	*mem_cached.Client
}

func (h *HandlerMemCachedSuite) flushAll() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", h.cfg.MemCached.Addr)
	if err != nil {
		h.T().Fatal(err)
	}
	nc, err := net.DialTimeout(tcpAddr.Network(), tcpAddr.String(), 500*time.Millisecond)
	if err != nil {
		h.T().Fatal(err)
	}
	err = nc.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		h.T().Fatal(err)
	}
	rw := bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc))

	if err != nil {
		h.T().Fatal(err)
	}
	_, err = rw.WriteString("flush_all")
	if err != nil {
		h.T().Fatal(err)
	}
	if err := rw.Flush(); err != nil {

		h.T().Fatal(err)
	}
}

func (h *HandlerMemCachedSuite) SetupTest() {
	var err error
	h.cfg, err = config.New(".env")
	if err != nil {
		log.Fatalf("Could not init config")
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "memcached",
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("11211/tcp")

	h.cfg.MemCached.Addr = hostAndPort
	connPool, err := mem_cached.NewConnPool(&mem_cached.TcpConfig{
		Addr:        h.cfg.MemCached.Addr,
		MaxIdleConn: h.cfg.MemCached.MaxIdleConn,
		MaxOpenConn: h.cfg.MemCached.MaxOpenConn,
	})
	if err != nil {
		h.T().Fatal("memCached initialization failed")
	}
	h.Client, err = mem_cached.NewClient(connPool)
	if err != nil {
		h.T().Fatal("memCached initialization failed")
	}
	h.RecordHandler = handler.NewHandler(h.Client)

}

func (h *HandlerMemCachedSuite) Test_ShouldAddAndDeleteManyRecords() {
	defer h.flushAll()
	wg := &sync.WaitGroup{}
	keys := make([]string, 1000)
	for i := 0; i < 10000; i++ {
		key := generate(i%10 + 1)
		keys = append(keys, key)
		wg.Add(1)
		go func(key string, wg *sync.WaitGroup) {
			defer wg.Done()
			_, err := h.RecordHandler.SetRecords(context.Background(), &proto.SetRecordsRequest{
				Record: &proto.Record{
					Key:   key,
					Value: key,
				},
			})
			assert.NoError(h.T(), err)
		}(key, wg)
	}
	wg.Wait()

	resp, err := h.RecordHandler.GetRecord(context.Background(), &proto.GetRecordRequest{Keys: keys})
	assert.NoError(h.T(), err)

	assert.Equal(h.T(), 10000, len(resp.Result))
}

func (h *HandlerMemCachedSuite) Test_ShouldAddAndGetRecord() {
	defer h.flushAll()
	_, err := h.RecordHandler.SetRecords(context.Background(), &proto.SetRecordsRequest{
		Record: &proto.Record{
			Key:   "key1",
			Value: "value1",
		},
	})
	assert.NoError(h.T(), err)
	resp, err := h.RecordHandler.GetRecord(context.Background(), &proto.GetRecordRequest{Keys: []string{"key1"}})
	assert.NoError(h.T(), err)
	expected := []*proto.Record{{
		Key:   "key1",
		Value: "value1",
	}}
	assert.Equal(h.T(), expected, resp.Result)
}

func (h *HandlerMemCachedSuite) Test_ShouldAddAndGetRecords() {
	defer h.flushAll()
	_, err := h.RecordHandler.SetRecords(context.Background(), &proto.SetRecordsRequest{
		Record: &proto.Record{
			Key:   "key1",
			Value: "value1",
		},
	})
	assert.NoError(h.T(), err)
	_, err = h.RecordHandler.SetRecords(context.Background(), &proto.SetRecordsRequest{
		Record: &proto.Record{
			Key:   "key2",
			Value: "value2",
		},
	})

	assert.NoError(h.T(), err)
	resp, err := h.RecordHandler.GetRecord(context.Background(), &proto.GetRecordRequest{Keys: []string{"key1", "key2"}})
	assert.NoError(h.T(), err)
	expected := []*proto.Record{{
		Key:   "key1",
		Value: "value1",
	},
		{
			Key:   "key2",
			Value: "value2",
		}}
	assert.Equal(h.T(), expected, resp.Result)
}

func (h *HandlerMemCachedSuite) Test_ShouldAddAndDeleteRecord() {
	defer h.flushAll()
	_, err := h.RecordHandler.SetRecords(context.Background(), &proto.SetRecordsRequest{
		Record: &proto.Record{
			Key:   "key1",
			Value: "value1",
		},
	})
	assert.NoError(h.T(), err)
	_, err = h.RecordHandler.DeleteRecord(context.Background(), &proto.DeleteRecordRequest{Key: "key1"})
	assert.NoError(h.T(), err)
	resp, err := h.RecordHandler.GetRecord(context.Background(), &proto.GetRecordRequest{Keys: []string{"key1"}})
	assert.NoError(h.T(), err)

	assert.Equal(h.T(), 0, len(resp.Result))
}

func generate(n int) string {
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
	str := make([]rune, n)
	for i := range str {
		str[i] = chars[rand.Intn(len(chars))]
	}
	return string(str)
}
func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerMemCachedSuite))
}
