package fabclient

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-protos-go/common"
	protomsp "github.com/hyperledger/fabric-protos-go/msp"
	protopeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspprovider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
)

type resourceManager interface {
	saveChannel(channelID, channelConfigPath string) error
	joinChannel(channelID string) error
	lifecycleInstallChaincode(chaincode Chaincode) (string, error)
	lifecycleApproveChaincode(channelID, packageID string, sequence int64, chaincode Chaincode) error
	lifecycleCheckChaincodeCommitReadiness(channelID, packageID string, sequence int64, chaincode Chaincode) bool
	lifecycleCommitChaincode(channelID string, sequence int64, chaincode Chaincode) error
	isLifecycleChaincodeInstalled(packageID string) bool
	isLifecycleChaincodeApproved(channelID, chaincodeName string, sequence int64) bool
}

type resourceManagementClient struct {
	adminIdentity mspprovider.SigningIdentity
	client        *resmgmt.Client
	defaultOpts   []resmgmt.RequestOption
	orderers      map[string]fab.OrdererConfig
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

	var randomOrderer string
	for orderer := range localContext.EndpointConfig().NetworkConfig().Orderers {
		randomOrderer = orderer
		break
	}

	rsmClient := &resourceManagementClient{
		adminIdentity: identity,
		client:        client,
		defaultOpts: []resmgmt.RequestOption{
			resmgmt.WithRetry(retry.DefaultResMgmtOpts),
			resmgmt.WithTargets(peers...),
			resmgmt.WithOrdererEndpoint(randomOrderer),
		},
		orderers: localContext.EndpointConfig().NetworkConfig().Orderers,
		peers:    peers,
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
			return fmt.Errorf("failed to save channel '%s': %w", channelID, err)
		}
	}

	return nil
}

func (rsm *resourceManagementClient) joinChannel(channelID string) error {
	err := rsm.client.JoinChannel(channelID, rsm.defaultOpts...)
	if err != nil && !strings.Contains(err.Error(), _channelAlreadyJoined) {
		return fmt.Errorf("failed to join channel '%s': %w", channelID, err)
	}

	return nil
}

func (rsm *resourceManagementClient) lifecycleInstallChaincode(chaincode Chaincode) (string, error) {
	label := chaincode.Name + "_" + chaincode.Version

	descriptor := &lifecycle.Descriptor{
		Path:  chaincode.Path,
		Type:  protopeer.ChaincodeSpec_GOLANG,
		Label: label,
	}

	chaincodePackage, err := lifecycle.NewCCPackage(descriptor)
	if err != nil {
		return "", fmt.Errorf("failed to install chaincode '%s': %w", chaincode.Name, err)
	}

	packageID := lifecycle.ComputePackageID(label, chaincodePackage)

	request := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: chaincodePackage,
	}

	res, err := rsm.client.LifecycleInstallCC(request, rsm.defaultOpts...)
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", fmt.Errorf("unexpected error occurred, failed to install chaincode '%s'", chaincode.Name)
	}

	for _, r := range res {
		if r.Status != 200 || r.PackageID != packageID {
			return "", fmt.Errorf("unexpected error occurred, failed to install chaincode '%s'", chaincode.Name)
		}
	}

	return packageID, nil
}

func (rsm *resourceManagementClient) lifecycleApproveChaincode(channelID, packageID string, sequence int64, chaincode Chaincode) error {
	request := resmgmt.LifecycleApproveCCRequest{
		Name:              chaincode.Name,
		Version:           chaincode.Version,
		PackageID:         packageID,
		Sequence:          sequence,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   generateChaincodePolicy(chaincode),
		InitRequired:      chaincode.InitRequired,
	}

	txID, err := rsm.client.LifecycleApproveCC(channelID, request, rsm.defaultOpts...)
	if err != nil {
		return fmt.Errorf("channel '%s', failed to approve chaincode '%s': %w", channelID, chaincode.Name, err)
	}

	if len(txID) == 0 {
		return fmt.Errorf("unexpected error occurred on channel '%s', failed to approve chaincode '%s'", channelID, chaincode.Name)
	}

	return nil
}

func (rsm *resourceManagementClient) lifecycleCheckChaincodeCommitReadiness(channelID, packageID string, sequence int64, chaincode Chaincode) bool {
	count := 0
	success := true
	checkChan := make(chan bool)
	numberOfPeers := len(rsm.peers)

	request := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              chaincode.Name,
		Version:           chaincode.Version,
		PackageID:         packageID,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   generateChaincodePolicy(chaincode),
		Sequence:          sequence,
		InitRequired:      chaincode.InitRequired,
	}

	for _, p := range rsm.peers {
		peer := p

		go func() {
			response, err := rsm.client.LifecycleCheckCCCommitReadiness(channelID, request, resmgmt.WithTargets(peer))
			if err != nil {
				checkChan <- false
				return
			}

			for _, org := range chaincode.MustBeApprovedByOrgs {
				ready, ok := response.Approvals[org]
				if !ok {
					checkChan <- false
					return
				}

				if !ready {
					checkChan <- false
					return
				}
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

func (rsm *resourceManagementClient) lifecycleCommitChaincode(channelID string, sequence int64, chaincode Chaincode) error {
	request := resmgmt.LifecycleCommitCCRequest{
		Name:              chaincode.Name,
		Version:           chaincode.Version,
		Sequence:          sequence,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   generateChaincodePolicy(chaincode),
		InitRequired:      chaincode.InitRequired,
	}

	txID, err := rsm.client.LifecycleCommitCC(channelID, request, rsm.defaultOpts...)
	if err != nil {
		return fmt.Errorf("channel '%s', failed to commit chaincode '%s': %w", channelID, chaincode.Name, err)
	}

	if len(txID) == 0 {
		return fmt.Errorf("unexpected error occurred on channel '%s', failed to commit chaincode '%s'", channelID, chaincode.Name)
	}

	return nil
}

func (rsm *resourceManagementClient) isLifecycleChaincodeInstalled(packageID string) bool {
	count := 0
	success := true
	checkChan := make(chan bool)
	numberOfPeers := len(rsm.peers)

	for _, p := range rsm.peers {
		peer := p

		go func() {
			response, err := rsm.client.LifecycleQueryInstalledCC(resmgmt.WithTargets(peer))
			if err != nil {
				checkChan <- false
				return
			}

			found := false
			for _, chaincodeInfo := range response {
				if chaincodeInfo.PackageID == packageID {
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

func (rsm *resourceManagementClient) isLifecycleChaincodeApproved(channelID, chaincodeName string, sequence int64) bool {
	count := 0
	success := true
	checkChan := make(chan bool)
	numberOfPeers := len(rsm.peers)

	request := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     chaincodeName,
		Sequence: sequence,
	}

	for _, p := range rsm.peers {
		peer := p

		go func() {
			response, err := rsm.client.LifecycleQueryApprovedCC(channelID, request, resmgmt.WithTargets(peer))
			if err != nil {
				checkChan <- false
				return
			}

			checkChan <- response.Name == chaincodeName && response.Sequence == sequence
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

func convertMSPRole(role string) protomsp.MSPRole_MSPRoleType {
	switch role {
	case "admin", "ADMIN":
		return protomsp.MSPRole_ADMIN
	case "client", "CLIENT":
		return protomsp.MSPRole_CLIENT
	case "member", "MEMBER":
		return protomsp.MSPRole_MEMBER
	case "orderer", "ORDERER":
		return protomsp.MSPRole_ORDERER
	case "peer", "PEER":
		return protomsp.MSPRole_PEER
	default:
		return -1
	}
}

func generateChaincodePolicy(chaincode Chaincode) *common.SignaturePolicyEnvelope {
	return policydsl.SignedByNOutOfGivenRole(
		int32(len(chaincode.MustBeApprovedByOrgs)),
		convertMSPRole(chaincode.Role),
		chaincode.MustBeApprovedByOrgs,
	)
}
