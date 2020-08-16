package fabclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type Client struct {
	config          *Config
	fabricSDK       *fabsdk.FabricSDK
	msp             membershipServiceProvider
	resourceManager resourceManager
}

func NewClientFromConfigFile(configPath string) (*Client, error) {
	cfg, err := NewConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}

	return NewClient(cfg)
}

func NewClient(cfg *Config) (*Client, error) {
	sdk, err := fabsdk.New(config.FromFile(cfg.SDKConfigPath))
	if err != nil {
		return nil, err
	}

	msp, err := newMembershipServiceProvider(cfg.Organization, sdk.Context())
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

	var client = &Client{
		config:          cfg,
		fabricSDK:       sdk,
		msp:             msp,
		resourceManager: rsm,
	}

	return client, nil
}

func (client *Client) AnchorPeerSetup(channel Channel) error {
	return client.resourceManager.anchorPeerSetup(channel)
}

func (client *Client) Config() *Config {
	return client.config
}

func (client *Client) CreateChannel(channel Channel) error {
	return client.resourceManager.createChannel(channel)
}

func (client *Client) JoinChannel(channel Channel) error {
	return client.resourceManager.joinChannel(channel)
}

func (client *Client) InstallChaincode(chaincode Chaincode) error {
	return client.resourceManager.installChaincode(chaincode)
}
