package fabclient

import (
	"errors"
	"fmt"
	"os"
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
	adminIdentity          mspprovider.SigningIdentity
	client                 *resmgmt.Client
	peers                  []fab.Peer
	withOrdererEndpointOpt resmgmt.RequestOption
	withRetryOpt           resmgmt.RequestOption
	withTargetPeersOpt     resmgmt.RequestOption
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

	var randomOrderer string
	for orderer := range localContext.EndpointConfig().NetworkConfig().Orderers {
		randomOrderer = orderer
		break
	}

	// for tests purpose
	if len(os.Getenv("TARGET_ORDERER")) > 0 {
		randomOrderer = os.Getenv("TARGET_ORDERER")
	}

	rsmClient := &resourceManagementClient{
		adminIdentity:          identity,
		client:                 client,
		peers:                  peers,
		withOrdererEndpointOpt: resmgmt.WithOrdererEndpoint(randomOrderer),
		withRetryOpt:           resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		withTargetPeersOpt:     resmgmt.WithTargets(peers...),
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

	if _, err := rsm.client.SaveChannel(request, rsm.withOrdererEndpointOpt, rsm.withRetryOpt); err != nil {
		if !strings.Contains(err.Error(), _channelAlreadyExists) {
			return fmt.Errorf("failed to save channel '%s': %w", channelID, err)
		}
	}

	return nil
}

func (rsm *resourceManagementClient) joinChannel(channelID string) error {
	err := rsm.client.JoinChannel(channelID, rsm.withOrdererEndpointOpt, rsm.withRetryOpt, rsm.withTargetPeersOpt)
	if err != nil && !strings.Contains(err.Error(), _channelAlreadyJoined) {
		return fmt.Errorf("failed to join channel '%s': %w", channelID, err)
	}

	return nil
}

func (rsm *resourceManagementClient) installChaincode(chaincode Chaincode) error {
	// let SDK determines GOPATH
	ccPackage, err := gopackager.NewCCPackage(chaincode.Path, "")
	if err != nil {
		return fmt.Errorf("failed to install chaincode '%s': %w", chaincode.Name, err)
	}

	request := resmgmt.InstallCCRequest{
		Name:    chaincode.Name,
		Path:    chaincode.Path,
		Version: chaincode.Version,
		Package: ccPackage,
	}

	res, err := rsm.client.InstallCC(request, rsm.withRetryOpt, rsm.withTargetPeersOpt)
	if err != nil {
		return fmt.Errorf("failed to install chaincode '%s': %w", chaincode.Name, err)
	}

	for _, r := range res {
		if r.Status != 200 {
			if err == nil {
				err = errors.New("unexpected error(s) occurred: ")
			}

			err = fmt.Errorf("%w\n[%s] failed to install chaincode '%s': %s", err, r.Target, chaincode.Name, r.Info)
		}
	}

	return err
}

func (rsm *resourceManagementClient) instantiateOrUpgradeChaincode(channelID string, chaincode Chaincode) error {
	if !rsm.isChaincodeInstalled(chaincode.Name) {
		return fmt.Errorf("chaincode '%s' has not been installed", chaincode.Name)
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
				done <- fmt.Errorf("[%s] failed to query instantiated chaincodes: %w", peer.URL(), err)
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
				done <- fmt.Errorf("[%s] failed to parse chaincode '%s' current version: %w", peer.URL(), chaincode.Name, err)
				return
			}

			newVersion, err := strconv.ParseFloat(chaincode.Version, 64)
			if err != nil {
				done <- fmt.Errorf("[%s] failed to parse chaincode '%s' new version: %w", peer.URL(), chaincode.Name, err)
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
						"following error(s) occurred when instantiating/upgrading chaincode '%s': ",
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
	policy, err := parsePolicy(chaincode.Policy)
	if err != nil {
		return fmt.Errorf("[%s] failed to instantiate chaincode '%s': %w", peer.URL(), chaincode.Name, err)
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

	if len(chaincode.Collections) > 0 {
		collectionsConfig, err := processChaincodeCollections(chaincode.Collections)
		if err != nil {
			return fmt.Errorf("[%s] failed to instantiate chaincode '%s': %w", peer.URL(), chaincode.Name, err)
		}

		request.CollConfig = collectionsConfig
	}

	response, err := rsm.client.InstantiateCC(channelID, request, resmgmt.WithTargets(peer))
	if err != nil {
		return fmt.Errorf("[%s] failed to instantiate chaincode '%s': %w", peer.URL(), chaincode.Name, err)
	}

	if len(response.TransactionID) == 0 {
		return fmt.Errorf("[%s] unexpected error occurred, failed to instantiate chaincode '%s'", peer.URL(), chaincode.Name)
	}

	return nil
}

func (rsm *resourceManagementClient) upgradeChaincode(channelID string, chaincode Chaincode, peer fab.Peer) error {
	policy, err := parsePolicy(chaincode.Policy)
	if err != nil {
		return fmt.Errorf("[%s] failed to upgrade chaincode '%s': %w", peer.URL(), chaincode.Name, err)
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

	if len(chaincode.Collections) > 0 {
		collectionsConfig, err := processChaincodeCollections(chaincode.Collections)
		if err != nil {
			return fmt.Errorf("[%s] failed to upgrade chaincode '%s': %w", peer.URL(), chaincode.Name, err)
		}

		request.CollConfig = collectionsConfig
	}

	response, err := rsm.client.UpgradeCC(channelID, request, resmgmt.WithTargets(peer))
	if err != nil {
		return fmt.Errorf("[%s] failed to upgrade chaincode '%s': %w", peer.URL(), chaincode.Name, err)
	}

	if len(response.TransactionID) == 0 {
		return fmt.Errorf("[%s] unexpected error occurred, failed to upgrade chaincode '%s'", peer.URL(), chaincode.Name)
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

			checkChan <- found
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

			checkChan <- found
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

func parsePolicy(policy string) (*common.SignaturePolicyEnvelope, error) {
	// if len(policy) == 0 {
	// 	return nil, nil
	// }

	signaturePolicyEnvelope, err := cauthdsl.FromString(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	return signaturePolicyEnvelope, nil
}

func processChaincodeCollections(collections []ChaincodeCollection) ([]*pb.CollectionConfig, error) {
	collectionsConfig := make([]*pb.CollectionConfig, 0, len(collections))

	for _, collection := range collections {
		policy, err := parsePolicy(collection.Policy)
		if err != nil {
			return nil, fmt.Errorf("failed to process configuration for collection '%s': %w", collection.Name, err)
		}

		collectionsConfig = append(collectionsConfig, &pb.CollectionConfig{
			Payload: &pb.CollectionConfig_StaticCollectionConfig{
				StaticCollectionConfig: &pb.StaticCollectionConfig{
					Name: collection.Name,
					MemberOrgsPolicy: &pb.CollectionPolicyConfig{
						Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
							SignaturePolicy: policy,
						},
					},
					RequiredPeerCount: collection.RequiredPeerCount,
					MaximumPeerCount:  collection.RequiredPeerCount,
					BlockToLive:       collection.BlockToLive,
					MemberOnlyRead:    collection.MemberOnlyRead,
				},
			},
		})
	}

	return collectionsConfig, nil
}
