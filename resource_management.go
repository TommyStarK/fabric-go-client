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
	// "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
)

type resourceManager interface {
	saveChannel(channelID, channelConfigPath string) error
	joinChannel(channelID string) error
	// lifecycleInstallChaincode(chaincode Chaincode) error
}

type resourceManagementClient struct {
	adminIdentity mspprovider.SigningIdentity
	client        *resmgmt.Client
	defaultOpts   []resmgmt.RequestOption
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

	rsmClient := &resourceManagementClient{
		adminIdentity: identity,
		client:        client,
		defaultOpts: []resmgmt.RequestOption{
			resmgmt.WithRetry(retry.DefaultResMgmtOpts),
			resmgmt.WithTargets(peers...),
		},
		peers: peers,
	}

	return rsmClient, nil
}

func (rsm *resourceManagementClient) saveChannel(channelID, channelConfigPath string) error {
	request := resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: channelConfigPath,
		SigningIdentities: []mspprovider.SigningIdentity{rsm.adminIdentity},
	}

	if _, err := rsm.client.SaveChannel(request, resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
		if !strings.Contains(err.Error(), _channelAlreadyExists) {
			return fmt.Errorf("failed to save channel %s: %w", channelID, err)
		}
	}

	return nil
}

func (rsm *resourceManagementClient) joinChannel(channelID string) error {
	err := rsm.client.JoinChannel(channelID, rsm.defaultOpts...)
	if err != nil && !strings.Contains(err.Error(), _channelAlreadyJoined) {
		return fmt.Errorf("failed to join channel %s: %w", channelID, err)
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
