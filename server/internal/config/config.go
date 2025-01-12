package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const configPath = "configs/server_config.json"

type Config struct {
	Host string `json:"host"`
	Port string `json:"port"`

	WriteTimeout string `json:"write_timeout"`
	ReadTimeout  string `json:"read_timeout"`

	ReadTimeoutDuration  time.Duration `json:"-"`
	WriteTimeoutDuration time.Duration `json:"-"`
}

func Load() (*Config, error) {
	cfg := Config{}

	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %w", err)
	}

	err = json.Unmarshal(fileBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	cfg.ReadTimeoutDuration, err = time.ParseDuration(cfg.ReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("cannot parse ReadTimeout, err=%w, ReadTimeout=%s", err, cfg.ReadTimeout)
	}

	cfg.WriteTimeoutDuration, err = time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("cannot parse WriteTimeout, err=%w, WriteTimeout=%s", err, cfg.WriteTimeout)
	}

	return &cfg, nil
}
