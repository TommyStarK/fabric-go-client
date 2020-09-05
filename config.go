package fabclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds the client configuration.
type Config struct {
	Chaincodes        []Chaincode `json:"chaincodes" yaml:"chaincodes"`
	Channels          []Channel   `json:"channels" yaml:"channels"`
	ConnectionProfile string      `json:"connectionProfile" yaml:"connectionProfile"`
	Identities        struct {
		Admin Identity   `json:"admin" yaml:"admin"`
		Users []Identity `json:"users" yaml:"users"`
	} `json:"identities" yaml:"identities"`
	Organization string `json:"organization" yaml:"organization"`
}

// NewConfigFromFile returns a new client configuration.
func NewConfigFromFile(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileAsBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var (
		config    = &Config{}
		extension = filepath.Ext(configPath)
	)

	switch extension {
	case ".json":
		if err := json.Unmarshal(fileAsBytes, config); err != nil {
			return nil, fmt.Errorf("failed to create client configuration from file %s: %w", configPath, err)
		}
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(fileAsBytes, config); err != nil {
			return nil, fmt.Errorf("failed to create client configuration from file %s: %w", configPath, err)
		}
	default:
		return nil, errors.New("invalid client configuration file extension, supported: .json|.yml|.yaml")
	}

	return config, nil
}
