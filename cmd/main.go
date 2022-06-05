package main

import (
	"github.com/romanzimoglyad/memcached/internal/config"
	"github.com/romanzimoglyad/memcached/internal/handler"
	"github.com/romanzimoglyad/memcached/internal/logger"
	in_memory "github.com/romanzimoglyad/memcached/internal/repository/in-memory"
	mem_cached "github.com/romanzimoglyad/memcached/internal/repository/mem-cached"
	"github.com/romanzimoglyad/memcached/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/ztrue/shutdown"
	"syscall"
)

func main() {
	cfg, err := config.New(".env")
	cfg.Print()
	if err != nil {
		log.Fatal().Err(err).Msgf("configuration failed")
	}
	logger.InitZeroLog(cfg)
	var r handler.Repository
	switch cfg.RepositoryType {
	case 1:
		connPool, err := mem_cached.NewConnPool(&mem_cached.TcpConfig{
			Addr:        cfg.MemCached.Addr,
			MaxIdleConn: cfg.MemCached.MaxIdleConn,
			MaxOpenConn: cfg.MemCached.MaxOpenConn,
		})
		defer connPool.Close()
		if err != nil {
			log.Panic().Err(err).Msgf("memCached initialization failed")
		}
		r, err = mem_cached.NewClient(connPool)
		if err != nil {
			log.Panic().Err(err).Msgf("memCached initialization failed")
		}
	case 2:
		r = in_memory.NewRecords()
	}
	handler := handler.NewHandler(r)
	server := server.NewServer(server.Conf{
		Handler: handler,
		Conf:    cfg,
	})
	server.Setup()
	server.Run()

	shutdown.Add(func() {
		log.Warn().Msgf("graceful shutdown: start")
		server.Shutdown()
		log.Info().Msgf("graceful shutdown: finished successful")
	})

	shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)
}
