package fabclient

import (
	"fmt"
	"strings"

	// protopeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspprovider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
)

type resourceManager interface {
	anchorPeerSetup(channel Channel) error
	createChannel(channel Channel) error
	joinChannel(channel Channel) error
	installChaincode(chaincode Chaincode) error
	// lifecycleInstallChaincode(chaincode Chaincode) error
}

type resourceManagementClient struct {
	adminIdentity mspprovider.SigningIdentity
	client        *resmgmt.Client
	peers         []fab.Peer
}

func newResourceManager(ctx context.ClientProvider, identity mspprovider.SigningIdentity) (resourceManager, error) {
	localContext, err := contextImpl.NewLocal(ctx)
	if err != nil {
		return nil, err
	}

	peers, err := localContext.LocalDiscoveryService().GetPeers()
	if err != nil {
		return nil, err
	}

	client, err := resmgmt.New(ctx)
	if err != nil {
		return nil, err
	}

	var rsmClient = &resourceManagementClient{
		adminIdentity: identity,
		client:        client,
		peers:         peers,
	}

	return rsmClient, nil
}

func (rsm *resourceManagementClient) anchorPeerSetup(channel Channel) error {
	var request = resmgmt.SaveChannelRequest{
		ChannelID:         channel.Name,
		ChannelConfigPath: channel.AnchorPeerConfigPath,
		SigningIdentities: []mspprovider.SigningIdentity{rsm.adminIdentity},
	}

	result, err := rsm.client.SaveChannel(request, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		if !strings.Contains(err.Error(), _channelAlreadyExists) {
			return fmt.Errorf("failed to setup anchor peer for channel %s: %w", channel.Name, err)
		}

		return nil
	}

	if len(result.TransactionID) == 0 {
		return fmt.Errorf("unexpected error occurred when setting up anchor peer for channel %s", channel.Name)
	}

	return nil
}

func (rsm *resourceManagementClient) createChannel(channel Channel) error {
	var request = resmgmt.SaveChannelRequest{
		ChannelID:         channel.Name,
		ChannelConfigPath: channel.ConfigPath,
		SigningIdentities: []mspprovider.SigningIdentity{rsm.adminIdentity},
	}

	result, err := rsm.client.SaveChannel(request, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		if !strings.Contains(err.Error(), _channelAlreadyExists) {
			return fmt.Errorf("failed to create channel %s: %w", channel.Name, err)
		}

		return nil
	}

	if len(result.TransactionID) == 0 {
		return fmt.Errorf("unexpected error occurred when creating channel %s", channel.Name)
	}

	return nil
}

func (rsm *resourceManagementClient) joinChannel(channel Channel) error {
	opts := make([]resmgmt.RequestOption, 0, 2)
	opts = append(opts, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	opts = append(opts, resmgmt.WithTargets(rsm.peers...))

	if err := rsm.client.JoinChannel(channel.Name, opts...); err != nil {
		if !strings.Contains(err.Error(), _channelAlreadyJoined) {
			return fmt.Errorf("failed to join channel %s: %w", channel.Name, err)
		}

		return nil
	}

	return nil
}

func (rsm *resourceManagementClient) installChaincode(chaincode Chaincode) error {
	// let SDK determines GOPATH
	ccPackage, err := gopackager.NewCCPackage(chaincode.Path, "")
	if err != nil {
		return err
	}

	var request = resmgmt.InstallCCRequest{
		Name:    chaincode.Name,
		Path:    chaincode.Path,
		Version: chaincode.Version,
		Package: ccPackage,
	}

	res, err := rsm.client.InstallCC(request, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return fmt.Errorf("unexpected error occurred, failed to install chaincode %s", chaincode.Name)
	}

	if res[0].Status != 200 {
		return fmt.Errorf("unexpected error occurred, failed to install chaincode %s", chaincode.Name)
	}

	return nil
}

// func (rsm *resourceManagementClient) lifecycleInstallChaincode(chaincode Chaincode) error {
// 	label := chaincode.Name + "_" + chaincode.Version

// 	descriptor := &lifecycle.Descriptor{
// 		Path:  chaincode.Path,
// 		Type:  protopeer.ChaincodeSpec_GOLANG,
// 		Label: label,
// 	}

// 	chaincodePackage, err := lifecycle.NewCCPackage(descriptor)
// 	if err != nil {
// 		return err
// 	}

// 	request := resmgmt.LifecycleInstallCCRequest{
// 		Label:   label,
// 		Package: chaincodePackage,
// 	}

// 	opts := make([]resmgmt.RequestOption, 0, 2)
// 	opts = append(opts, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
// 	opts = append(opts, resmgmt.WithTargets(rsm.peers...))

// 	res, err := rsm.client.LifecycleInstallCC(request, opts...)
// 	if err != nil {
// 		return err
// 	}

// 	if len(res) == 0 {
// 		return fmt.Errorf("unexpected error occurred, failed to lifecycle install chaincode %s", chaincode.Name)
// 	}

// 	if res[0].Status != 200 {
// 		return fmt.Errorf("unexpected error occurred, failed to lifecycle install chaincode %s", chaincode.Name)
// 	}

// 	return nil
// }
