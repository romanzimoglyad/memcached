package testing

import (
	"context"
	"github.com/romanzimoglyad/memcached/internal/config"
	"github.com/romanzimoglyad/memcached/internal/handler"
	in_memory "github.com/romanzimoglyad/memcached/internal/repository/in-memory"
	"github.com/romanzimoglyad/memcached/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type HandlerInMemorySuite struct {
	suite.Suite
	RecordHandler *handler.RecordHandler
	cfg           *config.Config
}

func (h *HandlerInMemorySuite) SetupTest() {
	var err error
	h.cfg, err = config.New(".env")
	if err != nil {
		log.Fatalf("Could not init config")
	}

	h.RecordHandler = handler.NewHandler(in_memory.NewRecords())

}
func (h *HandlerInMemorySuite) Test_ShouldAddAndGetRecord() {
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

func (h *HandlerInMemorySuite) Test_ShouldAddAndGetRecords() {
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

func (h *HandlerInMemorySuite) Test_ShouldAddAndDeleteRecord() {
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
func TestHandlerInMemorySuite(t *testing.T) {
	suite.Run(t, new(HandlerInMemorySuite))
}
