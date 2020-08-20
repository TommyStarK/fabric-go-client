package fabclient

import (
	"errors"
	"fmt"
	"sync"

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
	handlers    handlers
}

type channelsHandlers []channelHandlers

func (chls channelsHandlers) find(channelName string) handlers {
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

	mutex sync.RWMutex
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

	client := &Client{
		config:           cfg,
		fabricSDK:        sdk,
		msp:              msp,
		resourceManager:  rsm,
		channelsHandlers: make(channelsHandlers, 0, len(cfg.Channels)),
		mutex:            sync.RWMutex{},
	}

	return client, nil
}

// BindChannelToClient should be call once the peer joined a channel
func (client *Client) BindChannelToClient(channelID string) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if handlers := client.channelsHandlers.find(channelID); handlers != nil {
		return fmt.Errorf("channel %s already bound", channelID)
	}

	handlers := make(handlers, 0, len(client.config.Identities.Users))
	for _, user := range client.config.Identities.Users {
		userIdentity, err := client.msp.createSigningIdentity(user.Certificate, user.PrivateKey)
		if err != nil {
			return err
		}

		userContext := client.fabricSDK.ChannelContext(channelID, fabsdk.WithIdentity(userIdentity))

		chHandler, err := newChannelHandler(userContext)
		if err != nil {
			return err
		}

		handlers = append(handlers, handler{
			username: user.Username,
			handler:  chHandler,
		})
	}

	client.channelsHandlers = append(client.channelsHandlers, channelHandlers{
		channelName: channelID,
		handlers:    handlers,
	})

	return nil
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

func (client *Client) Invoke(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return nil, err
	}
	return handler.invoke(request, opts...)
}

func (client *Client) Query(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return nil, err
	}
	return handler.invoke(request, opts...)
}

func (client *Client) selectChannelHandler(opts ...Option) (channelHandler, error) {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

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

	if len(options.channelID) == 0 {
		if len(options.userIdentity) == 0 {
			return client.channelsHandlers[0].handlers[0].handler, nil
		}

		handler := client.channelsHandlers[0].handlers.find(options.userIdentity)
		if handler == nil {
			return nil, fmt.Errorf("no channel binding found for user context (%s)", options.userIdentity)
		}

		return handler, nil
	}

	chanHandlers := client.channelsHandlers.find(options.channelID)
	if chanHandlers == nil {
		return nil, fmt.Errorf("binding for channel %s not found", options.channelID)
	}

	if len(options.userIdentity) > 0 {
		handler := chanHandlers.find(options.userIdentity)
		if handler == nil {
			return nil, fmt.Errorf("no channel binding found for user context (%s)", options.userIdentity)
		}

		return handler, nil
	}

	return chanHandlers[0].handler, nil
}
