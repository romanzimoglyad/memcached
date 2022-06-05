package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"time"
)

func Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

	start := time.Now()

	resp, err = handler(ctx, req)
	log.Info().Msgf("Request: %s, Duration: %s", info.FullMethod, time.Since(start))
	if err != nil {
		log.Err(err).Msgf("Request: %s", info.FullMethod)
	}
	return resp, err
}
