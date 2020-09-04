package fabclient

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspprovider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

type resourceManager interface {
	saveChannel(channelID, channelConfigPath string) error
	joinChannel(channelID string) error
	installChaincode(chaincode Chaincode) error
	instantiateOrUpgradeChaincode(channelID string, chaincode Chaincode) error
	isChaincodeInstalled(chaincodeName string) bool
	isChaincodeInstantiated(channelID, chaincodeName, chaincodeVersion string) bool
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

var _ resourceManager = (*resourceManagementClient)(nil)

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

func (rsm *resourceManagementClient) installChaincode(chaincode Chaincode) error {
	// let SDK determines GOPATH
	ccPackage, err := gopackager.NewCCPackage(chaincode.Path, "")
	if err != nil {
		return err
	}

	request := resmgmt.InstallCCRequest{
		Name:    chaincode.Name,
		Path:    chaincode.Path,
		Version: chaincode.Version,
		Package: ccPackage,
	}

	res, err := rsm.client.InstallCC(request, rsm.defaultOpts...)
	if err != nil {
		return fmt.Errorf("failed to install chaincode %s (version: %s): %w", chaincode.Name, chaincode.Version, err)
	}

	for _, r := range res {
		if r.Status != 200 {
			return fmt.Errorf("peer (%s), unexpected error occurred, failed to install chaincode %s, %s", r.Target, chaincode.Name, r.Info)
		}
	}

	return nil
}

func (rsm *resourceManagementClient) instantiateOrUpgradeChaincode(channelID string, chaincode Chaincode) error {
	if !rsm.isChaincodeInstalled(chaincode.Name) {
		return fmt.Errorf("chaincode %s has not been installed", chaincode.Name)
	}

	var finalError error
	count := 0
	done := make(chan error)
	numberOfPeers := len(rsm.peers)

	for _, p := range rsm.peers {

		peer := p
		go func() {
			response, err := rsm.client.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargets(peer))
			if err != nil {
				done <- fmt.Errorf("peer (%s), failed to query instantiated chaincodes: %w", peer.URL(), err)
				return
			}

			witness := false
			var ChaincodeCurrentVersion string
			for _, chaincodeInfo := range response.Chaincodes {
				if chaincodeInfo.Name == chaincode.Name {
					witness = true
					ChaincodeCurrentVersion = chaincodeInfo.Version
				}
			}

			if !witness {
				done <- rsm.instantiateChaincode(channelID, chaincode, peer)
				return
			}

			currentVersion, err := strconv.ParseFloat(ChaincodeCurrentVersion, 64)
			if err != nil {
				done <- fmt.Errorf("peer (%s), failed to parse chaincode %s current version: %w", peer.URL(), chaincode.Name, err)
				return
			}

			newVersion, err := strconv.ParseFloat(chaincode.Version, 64)
			if err != nil {
				done <- fmt.Errorf("peer (%s), failed to parse chaincode %s new version: %w", peer.URL(), chaincode.Name, err)
				return
			}

			if newVersion > currentVersion {
				done <- rsm.upgradeChaincode(channelID, chaincode, peer)
				return
			}

			done <- nil
			return
		}()
	}

	for {
		select {
		case err := <-done:
			if err != nil {
				if finalError == nil {
					finalError = fmt.Errorf(
						"following error(s) occurred when instantiating/upgrading chaincode %s: ",
						chaincode.Name,
					)
				}

				finalError = fmt.Errorf("%s\n%w", finalError.Error(), err)
			}

			count++
			if count == numberOfPeers {
				close(done)
				return finalError
			}
		default:
		}
	}
}

func (rsm *resourceManagementClient) instantiateChaincode(channelID string, chaincode Chaincode, peer fab.Peer) error {
	policy, err := parseChaincodePolicy(chaincode.Policy)
	if err != nil {
		return err
	}

	request := resmgmt.InstantiateCCRequest{
		Name:    chaincode.Name,
		Path:    chaincode.Path,
		Version: chaincode.Version,
		Lang:    pb.ChaincodeSpec_GOLANG,
		Policy:  policy,
	}

	if len(chaincode.InitArgs) > 0 {
		request.Args = convertArrayOfStringsToArrayOfByteArrays(chaincode.InitArgs)
	}

	response, err := rsm.client.InstantiateCC(channelID, request, resmgmt.WithTargets(peer))
	if err != nil {
		return fmt.Errorf("peer (%s), failed to instantiate chaincode %s (version %s): %w", peer.URL(), chaincode.Name, chaincode.Version, err)
	}

	if len(response.TransactionID) == 0 {
		return fmt.Errorf("peer (%s), failed to instantiate chaincode %s (version %s)", peer.URL(), chaincode.Name, chaincode.Version)
	}

	return nil
}

func (rsm *resourceManagementClient) upgradeChaincode(channelID string, chaincode Chaincode, peer fab.Peer) error {
	policy, err := parseChaincodePolicy(chaincode.Policy)
	if err != nil {
		return err
	}

	request := resmgmt.UpgradeCCRequest{
		Name:    chaincode.Name,
		Path:    chaincode.Path,
		Version: chaincode.Version,
		Lang:    pb.ChaincodeSpec_GOLANG,
		Policy:  policy,
	}

	if len(chaincode.InitArgs) > 0 {
		request.Args = convertArrayOfStringsToArrayOfByteArrays(chaincode.InitArgs)
	}

	response, err := rsm.client.UpgradeCC(channelID, request, resmgmt.WithTargets(peer))
	if err != nil {
		return fmt.Errorf("peer (%s), failed to upgrade chaincode %s to version %s: %w", peer.URL(), chaincode.Name, chaincode.Version, err)
	}

	if len(response.TransactionID) == 0 {
		return fmt.Errorf("peer (%s), failed to upgrade chaincode %s to version %s", peer.URL(), chaincode.Name, chaincode.Version)
	}

	return nil
}

func (rsm *resourceManagementClient) isChaincodeInstalled(chaincodeName string) bool {
	count := 0
	checkChan := make(chan bool)
	numberOfPeers := len(rsm.peers)
	success := true

	for _, p := range rsm.peers {
		peer := p

		go func() {
			response, err := rsm.client.QueryInstalledChaincodes(resmgmt.WithTargets(peer))
			if err != nil {
				checkChan <- false
				return
			}

			found := false
			for _, chaincodeInfo := range response.Chaincodes {
				if chaincodeInfo.Name == chaincodeName {
					found = true
					break
				}
			}

			if !found {
				checkChan <- false
				return
			}

			checkChan <- true
			return
		}()
	}

	for {
		select {
		case witness := <-checkChan:
			if !witness {
				success = false
			}

			count++
			if count == numberOfPeers {
				close(checkChan)
				return success
			}
		default:
		}
	}
}

func (rsm *resourceManagementClient) isChaincodeInstantiated(channelID, chaincodeName, chaincodeVersion string) bool {
	count := 0
	checkChan := make(chan bool)
	numberOfPeers := len(rsm.peers)
	success := true

	for _, p := range rsm.peers {
		peer := p

		go func() {
			response, err := rsm.client.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargets(peer))
			if err != nil {
				checkChan <- false
				return
			}

			found := false
			for _, chaincodeInfo := range response.Chaincodes {
				if chaincodeInfo.Name == chaincodeName && chaincodeInfo.Version == chaincodeVersion {
					found = true
					break
				}
			}

			if !found {
				checkChan <- false
				return
			}

			checkChan <- true
			return
		}()
	}

	for {
		select {
		case witness := <-checkChan:
			if !witness {
				success = false
			}

			count++
			if count == numberOfPeers {
				close(checkChan)
				return success
			}
		default:
		}
	}
}

func parseChaincodePolicy(chaincodePolicy string) (*common.SignaturePolicyEnvelope, error) {
	var (
		err    error
		policy *common.SignaturePolicyEnvelope
	)

	if len(chaincodePolicy) > 0 {
		policy, err = cauthdsl.FromString(chaincodePolicy)
		if err != nil {
			return nil, err
		}
	}

	return policy, nil
}
