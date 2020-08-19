package fabclient

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type handler struct {
	username string
	handler  channelHandler
}

type handlers []handler

func (hls handlers) find(username string) channelHandler {
	for _, h := range hls {
		if h.username == username {
			return h.handler
		}
	}

	return nil
}

type channelHandlers struct {
	channelName string
	handlers    *handlers
}

type channelsHandlers []channelHandlers

func (chls channelsHandlers) find(channelName string) *handlers {
	for _, c := range chls {
		if c.channelName == channelName {
			return c.handlers
		}
	}

	return nil
}

type Client struct {
	config           *Config
	fabricSDK        *fabsdk.FabricSDK
	msp              membershipServiceProvider
	resourceManager  resourceManager
	channelsHandlers channelsHandlers
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

	chansHandlers := make(channelsHandlers, 0)
	for _, channel := range cfg.Channels {
		if chansHandlers.find(channel.Name) == nil {
			hls := make(handlers, 0)
			chansHandlers = append(chansHandlers, channelHandlers{
				channelName: channel.Name,
				handlers:    &hls,
			})
		}

		for _, user := range cfg.Identities.Users {
			chanHandlers := chansHandlers.find(channel.Name)
			if chanHandlers.find(user.Username) == nil {
				userIdentity, err := msp.createSigningIdentity(user.Certificate, user.PrivateKey)
				if err != nil {

				}

				userContext := sdk.ChannelContext(channel.Name, fabsdk.WithIdentity(userIdentity))

				chHandler, err := newChannelHandler(userContext)
				if err != nil {

				}

				*chanHandlers = append(*chanHandlers, handler{
					username: user.Username,
					handler:  chHandler,
				})
			}
		}
	}

	client := &Client{
		config:           cfg,
		fabricSDK:        sdk,
		msp:              msp,
		resourceManager:  rsm,
		channelsHandlers: chansHandlers,
	}

	return client, nil
}

func (client *Client) Config() *Config {
	return client.config
}

func (client *Client) SaveChannel(channelID, channelConfigPath string) error {
	return client.resourceManager.saveChannel(channelID, channelConfigPath)
}

func (client *Client) JoinChannel(channelID string) error {
	return client.resourceManager.joinChannel(channelID)
}

func (client *Client) InstallChaincode(chaincode Chaincode) error {
	return client.resourceManager.installChaincode(chaincode)
}

func (client *Client) InstantiateOrUpgradeChaincode(channelID string, chaincode Chaincode) error {
	return client.resourceManager.instantiateOrUpgradeChaincode(channelID, chaincode)
}

func (client *Client) IsChaincodeInstalled(chaincodeName string) bool {
	return client.resourceManager.isChaincodeInstalled(chaincodeName)
}

func (client *Client) IsChaincodeInstantiated(channelID, chaincodeName, chaincodeVersion string) bool {
	return client.resourceManager.isChaincodeInstantiated(channelID, chaincodeName, chaincodeVersion)
}

func (client *Client) selectChannelHandler(opts ...Option) (channelHandler, error) {
	options := &options{
		channelID:    "",
		userIdentity: "",
	}

	for _, opt := range opts {
		opt.apply(options)
	}

	if len(options.channelID) == 0 && len(client.channelsHandlers) > 1 {
		return nil, errors.New("no channel ID specified")
	}

	if len(options.channelID) == 0 && len(client.channelsHandlers) == 1 {
		if len(options.userIdentity) == 0 {
			handlers := *client.channelsHandlers[0].handlers
			return handlers[0].handler, nil
		}

		handlers := *client.channelsHandlers[0].handlers
		handler := handlers.find(options.userIdentity)
		if handler == nil {
			return nil, fmt.Errorf("handler")
		}

		return handler, nil
	}

	chanHandlers := client.channelsHandlers.find(options.channelID)
	if chanHandlers == nil {
		return nil, fmt.Errorf("handlers for channel %s not found", options.channelID)
	}

	if len(options.userIdentity) > 0 {
		handler := chanHandlers.find(options.userIdentity)
		if handler == nil {
			return nil, fmt.Errorf("handler for channel %s with user %s not found", options.channelID, options.userIdentity)
		}

		return handler, nil
	}

	handlers := *chanHandlers
	return handlers[0].handler, nil
}
