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
	saveChannel(channelID, channelConfigPath string) error
	joinChannel(channelID string) error
	installChaincode(chaincode Chaincode) error
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

	var rsmClient = &resourceManagementClient{
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
	var request = resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: channelConfigPath,
		SigningIdentities: []mspprovider.SigningIdentity{rsm.adminIdentity},
	}

	_, err := rsm.client.SaveChannel(request, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil && !strings.Contains(err.Error(), _channelAlreadyExists) {
		return fmt.Errorf("failed to save channel %s (%s): %w", channelID, channelConfigPath, err)
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

	if _, err := rsm.client.InstallCC(request, rsm.defaultOpts...); err != nil {
		return fmt.Errorf("failed to install chaincode %s (version: %s): %w", chaincode.Name, chaincode.Path, err)
	}

	success := true
	peers := make([]string, 0, len(rsm.peers))
	for _, peer := range rsm.peers {
		response, err := rsm.client.QueryInstalledChaincodes(resmgmt.WithTargets(peer))
		if err != nil {
			return err
		}

		found := false
		for _, chaincodeInfo := range response.Chaincodes {
			if chaincodeInfo.Name == chaincode.Name && chaincodeInfo.Version == chaincode.Version {
				found = true
				break
			}
		}

		if !found {
			success = false
			peers = append(peers, peer.URL())
		}
	}

	if !success {
		return fmt.Errorf("failed to install chaincode %s (version: %s) on peers (%s): %w",
			chaincode.Name, chaincode.Path, strings.Join(peers, ", "), err)
	}

	// if len(res) == 0 {
	// 	return fmt.Errorf("unexpected error occurred, failed to install chaincode %s", chaincode.Name)
	// }

	// if res[0].Status != 200 {
	// 	return fmt.Errorf("unexpected error occurred, failed to install chaincode %s", chaincode.Name)
	// }

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
