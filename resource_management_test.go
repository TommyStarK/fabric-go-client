package fabclient

import (
	"testing"
)

var ccPackageID string

func createUpdateAndJoinChannel(t *testing.T, client *Client) {
	channel := client.Config().Channels[0]

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err != nil {
		t.Fatal(err)
	}

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err != nil {
		t.Errorf("channel '%s' already exists, should not have returned an error: %w", channel.Name, err)
	}

	if err := client.SaveChannel(channel.Name, channel.AnchorPeerConfigPath); err != nil {
		t.Fatal(err)
	}

	if err = client.JoinChannel(channel.Name); err != nil {
		t.Fatal(err)
	}

	if err = client.JoinChannel(channel.Name); err != nil {
		t.Fatal(err)
	}
}

func channelManagementFailureCases(t *testing.T, client *Client) {
	channel := client.Config().Channels[0]
	channel.Name = "dummy"
	channel.ConfigPath = "/dummy"

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err == nil {
		t.Error("should have returned an error, we provided a wrong channel configuration path")
	}

	if err := client.JoinChannel(channel.Name); err == nil {
		t.Error("should have returned an error, channel 'dummy' does not exist")
	}
}

func installChaincodeContractAPI(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	channel := client.Config().Channels[0]

	chaincodeInitialVersionPackageID, err := client.LifecycleInstallChaincode(chaincode)
	if err != nil {
		t.Fatal(err)
	}

	ccPackageID = chaincodeInitialVersionPackageID
	if !client.IsLifecycleChaincodeInstalled(chaincodeInitialVersionPackageID) {
		t.Errorf("chaincode '%s' should be installed on all peers belonging to org MSP", chaincode.Name)
	}

	if err := client.LifecycleApproveChaincode(channel.Name, chaincodeInitialVersionPackageID, 1, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsLifecycleChaincodeApproved(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should be approved on channel %s", chaincode.Name, channel.Name)
	}

	if !client.LifecyleCheckChaincodeCommitReadiness(channel.Name, chaincodeInitialVersionPackageID, 1, chaincode) {
		t.Errorf("chaincode '%s' should be ready to be committed on channel %s", chaincode.Name, channel.Name)
	}

	// if err := client.LifecycleCommitChaincode(channel.Name, 1, chaincode); err != nil {
	// 	t.Fatal(err)
	// }
}

func chaincodeManagementFailureCases(t *testing.T, client *Client) {
	channel := client.Config().Channels[0]
	chaincode := client.Config().Chaincodes[0]
	chaincode.Name = "dummy"
	chaincode.Path = "/dummy"

	if _, err := client.LifecycleInstallChaincode(chaincode); err == nil {
		t.Error("should have returned an error, invalid chaincode path provided")
	}

	if client.IsLifecycleChaincodeInstalled("dummyChaincodePackageID") {
		t.Errorf("chaincode '%s' should not be installed on all peers belonging to org MSP", chaincode.Name)
	}

	if client.IsLifecycleChaincodeApproved(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should not be approved on channel %s", chaincode.Name, channel.Name)
	}

	cc := client.Config().Chaincodes[0]
	cc.MustBeApprovedByOrgs = append(cc.MustBeApprovedByOrgs, "Org2MSP")
	if client.LifecyleCheckChaincodeCommitReadiness(channel.Name, ccPackageID, 1, cc) {
		t.Errorf("chaincode '%s' should not be ready to be committed on channel %s", cc.Name, channel.Name)
	}

	// if err := client.LifecycleCommitChaincode(channel.Name, 1, chaincode); err == nil {
	// 	t.Error("should have returned an error, chaincode neither installed nor approved")
	// }
}
