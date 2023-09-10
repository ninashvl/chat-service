package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

func ParseAndValidate(filename string) (Config, error) {
	config := Config{}
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return Config{}, fmt.Errorf("decoding config file error: %v", err)
	}

	err := config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("config validation error: %v", err)
	}
	return config, nil
}
