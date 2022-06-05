package config

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
)

const (
	envFilename = ".env"
)

type Config struct {
	IP             string `env:"MCD_IP" envDefault:"0.0.0.0"`
	GRPCPort       string `env:"MCD_PORT" envDefault:"8801"`
	RepositoryType int    `env:"MCD_REPOSITORY_TYPE" envDefault:"1"` // 1 - memcached, 2- inmemory
	LogLevel       string `env:"MCD_LOGLEVEL" envDefault:"warn"`     // debug, info, warn, error, fatal, ""
	MemCached      *MemCached
}
type MemCached struct {
	Addr        string `env:"MCD_MEMCACHED_ADR" envDefault:"0.0.0.0:11211"`
	MaxOpenConn int    `env:"MCD_MAX_OPEN_CONN" envDefault:"500"`
	MaxIdleConn int    `env:"MCD_MAX_IDLE_CONN" envDefault:"500"`
}

func (c *Config) Print() {
	log.Info().Msgf("Config: %v", c)
	log.Info().Msgf("Config MemcaChed: %v", c.MemCached)
}

func New(path string) (*Config, error) {
	cfg := &Config{MemCached: &MemCached{}}

	if err := loadEnv(cfg, path); err != nil {
		return nil, err
	}

	return cfg, nil
}
func loadEnv(config interface{}, path string) error {
	if path == "" {
		path = envFilename
	}

	if fileExists(path) {
		err := godotenv.Load(path)
		if err != nil {
			return fmt.Errorf("error while loading existing %s file: %v", envFilename, err)
		}
	}

	if err := env.Parse(config); err != nil {
		return err
	}

	return nil
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
