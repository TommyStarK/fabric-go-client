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
	sdk, err := fabsdk.New(config.FromFile(cfg.ConnectionProfile))
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

// Config returns the client configuration.
func (client *Client) Config() *Config {
	config := &Config{
		Chaincodes:        make([]Chaincode, len(client.config.Chaincodes)),
		Channels:          make([]Channel, len(client.config.Channels)),
		ConnectionProfile: client.config.ConnectionProfile,
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
		Organization: client.config.Organization,
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

// JoinChannel allows for peers to join existing channel.
func (client *Client) JoinChannel(channelID string) error {
	return client.resourceManager.joinChannel(channelID)
}

// LifecycleInstallChaincode installs a chaincode package using Fabric 2.0 chaincode lifecycle.
// It returns the chaincode package ID if the install succeeded.
func (client *Client) LifecycleInstallChaincode(chaincode Chaincode) (string, error) {
	return client.resourceManager.lifecycleInstallChaincode(chaincode)
}

// LifecycleApproveChaincode approves a chaincode for an organization.
func (client *Client) LifecycleApproveChaincode(channelID, packageID string, sequence int64, chaincode Chaincode) error {
	return client.resourceManager.lifecycleApproveChaincode(channelID, packageID, sequence, chaincode)
}

// LifecyleCheckChaincodeCommitReadiness returns wheter the given chaincode is ready to be committed on the specified channel.
func (client *Client) LifecyleCheckChaincodeCommitReadiness(channelID, packageID string, sequence int64, chaincode Chaincode) bool {
	return client.resourceManager.lifecycleCheckChaincodeCommitReadiness(channelID, packageID, sequence, chaincode)
}

// LifecycleCommitChaincode commits the chaincode to the given channel.
func (client *Client) LifecycleCommitChaincode(channelID string, sequence int64, chaincode Chaincode) error {
	return client.resourceManager.lifecycleCommitChaincode(channelID, sequence, chaincode)
}

// IsLifecycleChaincodeInstalled returns whether the given chaincode has been installed or not using the chaincode package ID.
func (client *Client) IsLifecycleChaincodeInstalled(packageID string) bool {
	return client.resourceManager.isLifecycleChaincodeInstalled(packageID)
}

// IsLifecycleChaincodeApproved returns whether the given chaincode has been approved on the specified channel.
func (client *Client) IsLifecycleChaincodeApproved(channelID, chaincodeName string, sequence int64) bool {
	return client.resourceManager.isLifecycleChaincodeApproved(channelID, chaincodeName, sequence)
}
