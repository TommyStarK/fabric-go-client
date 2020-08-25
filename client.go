package fabclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// Client enables:
// - managing resources in Fabric network
// - access to a channel on a Fabric network.
// - ledger queries on a Fabric network.
type Client struct {
	config          *Config
	fabricSDK       *fabsdk.FabricSDK
	msp             membershipServiceProvider
	resourceManager resourceManager
}

// NewClientFromConfigFile returns a client instance from a config file.
func NewClientFromConfigFile(configPath string) (*Client, error) {
	cfg, err := NewConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}

	return NewClient(cfg)
}

// NewClient returns a Client instance.
func NewClient(cfg *Config) (*Client, error) {
	sdk, err := fabsdk.New(config.FromFile(cfg.SDKConfigPath))
	if err != nil {
		return nil, err
	}

	msp, err := newMembershipServiceProvider(sdk.Context(), cfg.Organization)
	if err != nil {
		return nil, err
	}

	adminIdentity, err := msp.createSigningIdentity(cfg.Identities.Admin.Certificate, cfg.Identities.Admin.PrivateKey)
	if err != nil {
		return nil, err
	}

	adminContext := sdk.Context(fabsdk.WithIdentity(adminIdentity), fabsdk.WithOrg(cfg.Organization))

	rsm, err := newResourceManager(adminContext, adminIdentity)
	if err != nil {
		return nil, err
	}

	client := &Client{
		config:          cfg,
		fabricSDK:       sdk,
		msp:             msp,
		resourceManager: rsm,
	}

	return client, nil
}

// Config returns the client configuration
func (client *Client) Config() *Config {
	config := &Config{
		Chaincodes: make([]Chaincode, len(client.config.Chaincodes)),
		Channels:   make([]Channel, len(client.config.Channels)),
		Identities: struct {
			Admin Identity   `json:"admin" yaml:"admin"`
			Users []Identity `json:"users" yaml:"users"`
		}{
			Admin: Identity{
				Certificate: client.config.Identities.Admin.Certificate,
				PrivateKey:  client.config.Identities.Admin.PrivateKey,
				Username:    client.config.Identities.Admin.Username,
			},
			Users: make([]Identity, len(client.config.Identities.Users)),
		},
		Organization:  client.config.Organization,
		Version:       client.config.Version,
		SDKConfigPath: client.config.SDKConfigPath,
	}

	copy(config.Identities.Users, client.config.Identities.Users)
	copy(config.Chaincodes, client.config.Chaincodes)
	copy(config.Channels, client.config.Channels)
	return config
}

// SaveChannel creates or updates channel.
func (client *Client) SaveChannel(channelID, channelConfigPath string) error {
	return client.resourceManager.saveChannel(channelID, channelConfigPath)
}

// JoinChannel allows for peers to join existing channel
func (client *Client) JoinChannel(channelID string) error {
	return client.resourceManager.joinChannel(channelID)
}
