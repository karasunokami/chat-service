package config

import (
	"fmt"
	"os"

	"github.com/karasunokami/chat-service/internal/validator"

	"github.com/pelletier/go-toml"
)

func ParseAndValidate(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("os read file, filepath=%s, err=%w", filename, err)
	}

	var cfg Config

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("toml unmarshal, err=%v", err)
	}

	err = validator.Validator.Struct(cfg)
	if err != nil {
		return Config{}, fmt.Errorf("validate, err=%v", err)
	}

	return cfg, nil
}
