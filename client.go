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

// Client API enables to manage resources in a Fabric network, access to a channel,
// perform chaincode related operations.
type Client struct {
	config           *Config
	fabricSDK        *fabsdk.FabricSDK
	msp              membershipServiceProvider
	resourceManager  resourceManager
	channelsHandlers channelsHandlers

	mutex sync.RWMutex
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
		config:           cfg,
		fabricSDK:        sdk,
		msp:              msp,
		resourceManager:  rsm,
		channelsHandlers: make(channelsHandlers, 0, len(cfg.Channels)),
		mutex:            sync.RWMutex{},
	}

	return client, nil
}

func (client *Client) bindChannel(channelID string) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if handlers := client.channelsHandlers.find(channelID); handlers != nil {
		return nil
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
			return fmt.Errorf("failed to bind channel '%s' to client: %w", channelID, err)
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
	if err := client.resourceManager.joinChannel(channelID); err != nil {
		return err
	}

	return client.bindChannel(channelID)
}

// InstallChaincode allows administrators to install chaincode onto the filesystem of a peer.
func (client *Client) InstallChaincode(chaincode Chaincode) error {
	return client.resourceManager.installChaincode(chaincode)
}

// InstantiateOrUpgradeChaincode instantiates or upgrades chaincode.
func (client *Client) InstantiateOrUpgradeChaincode(channelID string, chaincode Chaincode) error {
	return client.resourceManager.instantiateOrUpgradeChaincode(channelID, chaincode)
}

// IsChaincodeInstalled returns whether the given chaincode has been installed or not.
func (client *Client) IsChaincodeInstalled(chaincodeName string) bool {
	return client.resourceManager.isChaincodeInstalled(chaincodeName)
}

// IsChaincodeInstantiated returns whether the given chaincode has been instantiated or not.
func (client *Client) IsChaincodeInstantiated(channelID, chaincodeName, chaincodeVersion string) bool {
	return client.resourceManager.isChaincodeInstantiated(channelID, chaincodeName, chaincodeVersion)
}

// Invoke prepares and executes transaction using request and optional request options.
func (client *Client) Invoke(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return nil, err
	}
	return handler.invoke(request, opts...)
}

// Query chaincode using request and optional request options.
func (client *Client) Query(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return nil, err
	}
	return handler.query(request, opts...)
}

// QueryBlockByTxID queries for block which contains a transaction.
func (client *Client) QueryBlockByTxID(txID string, opts ...Option) (*Block, error) {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return nil, err
	}
	return handler.queryBlockByTxID(txID)
}

// RegisterChaincodeEvent registers for chaincode events. Unregister must be called when the registration is no longer needed.
func (client *Client) RegisterChaincodeEvent(chaincodeID, eventFilter string, opts ...Option) (<-chan *ChaincodeEvent, error) {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return nil, err
	}
	return handler.registerChaincodeEvent(chaincodeID, eventFilter)
}

// UnregisterChaincodeEvent removes the given registration and closes the event channel.
func (client *Client) UnregisterChaincodeEvent(eventFilter string, opts ...Option) error {
	handler, err := client.selectChannelHandler(opts...)
	if err != nil {
		return err
	}
	handler.unregisterChaincodeEvent(eventFilter)
	return nil
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
		return nil, errors.New("cannot determine channel context, multiple channels bound to client")
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
		return nil, fmt.Errorf("binding for channel '%s' not found", options.channelID)
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
