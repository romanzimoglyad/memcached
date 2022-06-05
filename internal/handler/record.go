package handler

import (
	"context"
	"github.com/romanzimoglyad/memcached/internal/model"
	"github.com/romanzimoglyad/memcached/proto"
)

type RecordHandler struct {
	proto.UnimplementedRecordServiceServer
	repository Repository
}

func NewHandler(repository Repository) *RecordHandler {
	return &RecordHandler{repository: repository}
}

var _ proto.RecordServiceServer = (*RecordHandler)(nil)

func (r *RecordHandler) GetRecord(ctx context.Context, request *proto.GetRecordRequest) (*proto.GetRecordResponse, error) {
	out := &proto.GetRecordResponse{}
	records, err := r.repository.Get(request.Keys...)
	if err != nil {
		return out, err
	}
	pbRecords := make([]*proto.Record, len(records))
	for i := range records {
		pbRecords[i] = &proto.Record{
			Key:   records[i].Key,
			Value: records[i].Value,
		}
	}
	out.Result = pbRecords
	return out, nil
}

func (r *RecordHandler) SetRecords(ctx context.Context, request *proto.SetRecordsRequest) (*proto.Empty, error) {

	return &proto.Empty{}, r.repository.Set(model.Record{
		Key:   request.Record.Key,
		Value: request.Record.Value,
	})
}

func (r *RecordHandler) DeleteRecord(ctx context.Context, request *proto.DeleteRecordRequest) (*proto.Empty, error) {
	return &proto.Empty{}, r.repository.Delete(request.Key)
}
