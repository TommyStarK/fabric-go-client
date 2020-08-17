package fabclient

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Chaincode struct {
	InitArgs []string `json:"initArgs" yaml:"initArgs"`
	Name     string   `json:"name" yaml:"name"`
	Path     string   `json:"path" yaml:"path"`
	Policy   string   `json:"policy,omitempty" yaml:"policy,omitempty"`
	Version  string   `json:"version" yaml:"version"`
}

type Channel struct {
	AnchorPeerConfigPath string `json:"anchorPeerConfigPath,omitempty" yaml:"anchorPeerConfigPath,omitempty"`
	ConfigPath           string `json:"configPath" yaml:"configPath"`
	Name                 string `json:"name" yaml:"name"`
}

type Identity struct {
	Certificate string `json:"certificate" yaml:"certificate"`
	PrivateKey  string `json:"privateKey" yaml:"privateKey"`
	Username    string `json:"username" yaml:"username"`
}

type Identities struct {
	Admin Identity   `json:"admin" yaml:"admin"`
	Users []Identity `json:"users" yaml:"users"`
}

type Config struct {
	Chaincodes    []Chaincode `json:"chaincodes" yaml:"chaincodes"`
	Channels      []Channel   `json:"channels" yaml:"channels"`
	Identities    Identities  `json:"identities" yaml:"identities"`
	Organization  string      `json:"organization" yaml:"organization"`
	SDKConfigPath string      `json:"sdkConfigPath" yaml:"sdkConfigPath"`
	Version       string      `json:"version" yaml:"version"`
}

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
			return nil, err
		}
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(fileAsBytes, config); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid client configuration file extension, supported: .json .yml .yaml")
	}

	return config, nil
}
