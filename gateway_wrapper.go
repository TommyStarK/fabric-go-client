package fabclient

import (
	"errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// A Contract object represents a smart contract instance in a network.
// Applications should get a Contract instance from a Network using the GetContract method
type Contract struct {
	contract *gateway.Contract
}

// EvaluateTransaction will evaluate a transaction function and return its results.
// The transaction function 'name' will be evaluated on the endorsing peers but the responses will not be sent
// to the ordering service and hence will not be committed to the ledger. This can be used for querying the world state.
func (c *Contract) EvaluateTransaction(name string, args []string) ([]byte, error) {
	return c.contract.EvaluateTransaction(name, args...)
}

// Name returns the name of the smart contract.
func (c *Contract) Name() string {
	return c.contract.Name()
}

// SubmitTransaction will submit a transaction to the ledger. The transaction function 'name' will be evaluated on
// the endorsing peers and then submitted to the ordering service for committing to the ledger.
func (c *Contract) SubmitTransaction(name string, args []string) ([]byte, error) {
	return c.contract.SubmitTransaction(name, args...)
}

// Gateway is the entry point to a Fabric network
type Gateway struct {
	gateway *gateway.Gateway
}

// Connect to a gateway defined by a network config file. Must specify a config option, an identity option.
func Connect(config GatewayConfigOption, identity WalletIdentityOption) (*Gateway, error) {
	g, err := gateway.Connect(config, identity)
	if err != nil {
		return nil, err
	}

	gw := &Gateway{gateway: g}
	return gw, nil
}

// Close the gateway connection and all associated resources, including removing listeners attached to networks and contracts
// created by the gateway.
func (gw *Gateway) Close() {
	gw.gateway.Close()
}

// GetNetwork returns an object representing a network channel.
func (gw *Gateway) GetNetwork(name string) (*Network, error) {
	net, err := gw.gateway.GetNetwork(name)
	if err != nil {
		return nil, err
	}

	network := &Network{network: net}
	return network, nil
}

// GatewayConfigOption specifies the gateway configuration source.
type GatewayConfigOption gateway.ConfigOption

// WithConfigFromFile configures the gateway from a network config file.
func WithConfigFromFile(configFile string) GatewayConfigOption {
	return gateway.WithConfig(config.FromFile(configFile))
}

// A Network object represents the set of peers in a Fabric network (channel).
// Applications should get a Network instance from a Gateway using the GetNetwork method.
type Network struct {
	network *gateway.Network
}

// GetContract returns instance of a smart contract on the current network.
func (net *Network) GetContract(chaincodeID string) *Contract {
	contract := &Contract{contract: net.network.GetContract(chaincodeID)}
	return contract
}

// Name is the name of the network (also known as channel name).
func (net *Network) Name() string {
	return net.network.Name()
}

// Wallet stores identity information used to connect to a Hyperledger Fabric network.
type Wallet struct {
	wallet *gateway.Wallet
}

// NewFileSystemWallet creates an instance of a wallet.
func NewFileSystemWallet(path string) (*Wallet, error) {
	wallet, err := gateway.NewFileSystemWallet(path)
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		wallet: wallet,
	}

	return w, nil
}

// Exists tests whether the wallet contains an identity for the given label.
func (w *Wallet) Exists(label string) bool {
	return w.wallet.Exists(label)
}

// Get an identity from the wallet.
func (w *Wallet) Get(label string) (WalletIdentity, error) {
	id, err := w.wallet.Get(label)
	if err != nil {
		return nil, err
	}

	if _, ok := id.(WalletIdentity); !ok {
		return nil, errors.New("failed to retrieve WalletIdentity")
	}

	return id.(WalletIdentity), nil
}

// List returns the labels of all identities in the wallet.
func (w *Wallet) List() ([]string, error) {
	return w.wallet.List()
}

// Put an identity into the wallet.
func (w *Wallet) Put(label string, id WalletIdentity) error {
	return w.wallet.Put(label, id)
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
func (w *Wallet) Remove(label string) error {
	return w.wallet.Remove(label)
}

// WalletIdentity represents a specific identity format.
type WalletIdentity gateway.Identity

// WalletIdentityOption specifies the user identity under which all transactions are performed for this gateway instance.
type WalletIdentityOption gateway.IdentityOption

// WithIdentity is an optional argument to the Connect method which specifies the identity that is to be used to connect to the network.
// All operations under this gateway connection will be performed using this identity.
func WithIdentity(wallet *Wallet, label string) WalletIdentityOption {
	return gateway.WithIdentity(wallet.wallet, label)
}

// WalletX509Identity represents an X509 identity.
type WalletX509Identity struct {
	*gateway.X509Identity
}

// NewWalletX509Identity creates an X509 identity for storage in a wallet.
func NewWalletX509Identity(mspid string, cert string, key string) *WalletX509Identity {
	return &WalletX509Identity{gateway.NewX509Identity(mspid, cert, key)}
}
