package fabclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// Client API enables to manage resources in a Fabric network, access to a channel,
// perform chaincode related operations.
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

// Close frees up caches and connections being maintained by the SDK.
func (client *Client) Close() {
	client.fabricSDK.Close()
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

// LifecycleInstallChaincode installs a chaincode package using Fabric 2.0 chaincode lifecycle. Returns the chaincode package ID if the install succeeded.
func (client *Client) LifecycleInstallChaincode(chaincode Chaincode) (string, error) {
	return client.resourceManager.lifecycleInstallChaincode(chaincode)
}

// LifecycleApproveChaincode approves a chaincode for an organization.
func (client *Client) LifecycleApproveChaincode(channelID, packageID string, sequence int64, chaincode Chaincode) error {
	return client.resourceManager.lifecycleApproveChaincode(channelID, packageID, sequence, chaincode)
}

// LifecyleCheckChaincodeCommitReadiness checks the 'commit readiness' of a chaincode. Returns a map holding the org approvals.
func (client *Client) LifecyleCheckChaincodeCommitReadiness(channelID, packageID string, sequence int64, chaincode Chaincode) (map[string]bool, error) {
	return client.resourceManager.lifecycleCheckChaincodeCommitReadiness(channelID, packageID, sequence, chaincode)
}

// LifecycleCommitChaincode commits the chaincode to the given channel.
func (client *Client) LifecycleCommitChaincode(channelID string, sequence int64, chaincode Chaincode) error {
	return client.resourceManager.lifecycleCommitChaincode(channelID, sequence, chaincode)
}

// IsChaincodeInstalled returns whether the given chaincode has been installed or not.
func (client *Client) IsChaincodeInstalled(packageID string) bool {
	return client.resourceManager.isChaincodeInstalled(packageID)
}

// IsChaincodeApproved returns whether the given chaincode has been approved or not.
func (client *Client) IsChaincodeApproved(channelID, chaincodeName string, sequence int64) bool {
	return client.resourceManager.isChaincodeApproved(channelID, chaincodeName, sequence)
}

// IsChaincodeCommitted returns whether the given chaincode has been committed or not.
func (client *Client) IsChaincodeCommitted(channelID, chaincodeName string, sequence int64) bool {
	return client.resourceManager.isChaincodeCommitted(channelID, chaincodeName, sequence)
}
