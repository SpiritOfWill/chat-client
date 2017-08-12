package config

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

// Type - main config
type Type struct {
	Host         string        `default:"localhost" toml:"host"`
	Port         uint          `default:"8080" toml:"port"`
	MaxChunkSize int           `default:"1000000" toml:"max_chunk_size"`
	Poll         time.Duration `default:"1" toml:"poll_interval"`
}

const configPath = "config.toml"

// Conf - global config
var Conf Type

func init() {
	if _, err := toml.DecodeFile(configPath, &Conf); err != nil {
		log.Panicf("failed to read config: %s", err)
	}
}
