package server

import (
	"context"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/romanzimoglyad/memcached/internal/config"
	"github.com/romanzimoglyad/memcached/internal/handler"
	"github.com/romanzimoglyad/memcached/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"runtime/debug"
)

type Server struct {
	server  *grpc.Server
	handler *handler.RecordHandler
	conf    *config.Config
}

type Conf struct {
	Handler *handler.RecordHandler
	Conf    *config.Config
}

func NewServer(serverConf Conf) *Server {
	return &Server{
		handler: serverConf.Handler,
		conf:    serverConf.Conf,
	}
}
func (s *Server) Setup() {
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		Intercept,
		grpcRecovery.UnaryServerInterceptor(
			grpcRecovery.WithRecoveryHandlerContext(RecoveryHandlerFuncContext),
		),
	}

	s.server = grpc.NewServer(
		grpcMiddleware.WithUnaryServerChain(unaryInterceptors...),
	)
	proto.RegisterRecordServiceServer(s.server, s.handler)

}

func (s *Server) Run() {
	log.Info().Msg("Grpc server starts...")

	go func() {
		listen, err := net.Listen("tcp", net.JoinHostPort(s.conf.IP, s.conf.GRPCPort))
		if err != nil {
			log.Fatal().Err(err).Msgf("grpc: failed to listen")
		}

		log.Info().Msgf("grpc: server listen %s:%s", s.conf.IP, s.conf.GRPCPort)

		if err = s.server.Serve(listen); err != nil {
			log.Fatal().Err(err).Msgf("grpc: failed to start server")
		}

	}()
}

// Shutdown performs graceful shutdown sequence for server instance.
func (s *Server) Shutdown() {
	if s.server == nil {
		return
	}

	log.Info().Msgf("grpc: closing connections...")

	s.server.GracefulStop()
}
func RecoveryHandlerFuncContext(_ context.Context, p interface{}) (err error) {
	log.Debug().Msgf("panic: %s", debug.Stack())
	return status.Errorf(codes.Internal, "%v", p)
}
