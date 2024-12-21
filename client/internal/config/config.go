package config

import (
	"encoding/json"
	"log"
	"os"
)

const configPath = "configs/client_config.json"

type Config struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func MustLoad() Config {
	cfg := Config{}

	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("cannot open config file: %v", err)
	}

	err = json.Unmarshal(fileBytes, &cfg)
	if err != nil {
		log.Fatalf("cannot unmarshal config: %v", err)
	}

	return cfg
}
