package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Addr         string `yaml:"addr"`
		SinglePlayer bool   `yaml:"single_player"`
		Debug        bool   `yaml:"debug"`
	} `yaml:"server"`
}

const configFilePath = "configs/config.yaml"

var appConfig *Config

func init() {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("failed to read config file at %s: %v", configFilePath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	appConfig = &cfg
}

func Get() *Config {
	if appConfig == nil {
		log.Fatalf("config not properly initialized")
	}

	return appConfig
}
