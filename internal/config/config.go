package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// InitYamlCfg reads any type of config (usign Go generics) from YAML file.
// It can be used either server and client sides for different types of config.
func InitYamlCfg[ConfType any](envConfigPath string) (*ConfType, error) {
	configPath, ok := os.LookupEnv(envConfigPath)
	if !ok {
		return nil, errors.New(envConfigPath + " env var is not set")
	}

	rawYAML, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.WithMessage(err, "reading config file")
	}

	var cfg ConfType
	err = yaml.Unmarshal(rawYAML, &cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing yaml")
	}

	return &cfg, nil
}
