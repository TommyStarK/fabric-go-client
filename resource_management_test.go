package fabclient

import (
	"testing"
)

func org1CreateUpdateAndJoinChannel(t *testing.T, client *Client) {
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

	if err := client.JoinChannel(channel.Name); err != nil {
		t.Fatal(err)
	}
}

func org2UpdateAndJoinChannel(t *testing.T, client *Client) {
	channel := client.Config().Channels[0]

	if err := client.SaveChannel(channel.Name, channel.AnchorPeerConfigPath); err != nil {
		t.Fatal(err)
	}

	if err := client.JoinChannel(channel.Name); err != nil {
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

func org1InstallAndApproveChaincodeContractAPI(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	channel := client.Config().Channels[0]

	chaincodeInitialVersionPackageID, err := client.LifecycleInstallChaincode(chaincode)
	if err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstalled(chaincodeInitialVersionPackageID) {
		t.Errorf("chaincode '%s' should be installed on all peers belonging to org MSP", chaincode.Name)
	}

	if err := client.LifecycleApproveChaincode(channel.Name, chaincodeInitialVersionPackageID, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeApproved(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should be approved on channel '%s'", chaincode.Name, channel.Name)
	}

	res, err := client.LifecyleCheckChaincodeCommitReadiness(channel.Name, chaincodeInitialVersionPackageID, chaincode)
	if err != nil {
		t.Errorf("chaincode '%s' should be ready to be committed on channel '%s'", chaincode.Name, channel.Name)
	}

	ready, ok := res["Org1MSP"]
	if !ok || !ready {
		t.Errorf("chaincode '%s' should be ready to be committed for Org1MSP", chaincode.Name)
	}
}

func org2InstallAndApproveChaincodeContractAPI(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	channel := client.Config().Channels[0]

	chaincodeInitialVersionPackageID, err := client.LifecycleInstallChaincode(chaincode)
	if err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstalled(chaincodeInitialVersionPackageID) {
		t.Errorf("chaincode '%s' should be installed on all peers belonging to org MSP", chaincode.Name)
	}

	if err := client.LifecycleApproveChaincode(channel.Name, chaincodeInitialVersionPackageID, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeApproved(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should be approved on channel '%s'", chaincode.Name, channel.Name)
	}

	res, err := client.LifecyleCheckChaincodeCommitReadiness(channel.Name, chaincodeInitialVersionPackageID, chaincode)
	if err != nil {
		t.Errorf("chaincode '%s' should be ready to be committed on channel '%s'", chaincode.Name, channel.Name)
	}

	ready, ok := res["Org2MSP"]
	if !ok || !ready {
		t.Errorf("chaincode '%s' should be ready to be committed for Org2MSP", chaincode.Name)
	}
}

func org1CommitChaincode(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	channel := client.Config().Channels[0]

	if err := client.LifecycleCommitChaincode(channel.Name, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeCommitted(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should be committed on channel '%s'", chaincode.Name, channel.Name)
	}
}

func chaincodeManagementFailureCases(t *testing.T, client *Client) {
	channel := client.Config().Channels[0]
	chaincode := client.Config().Chaincodes[0]
	chaincode.Name = "dummy"
	chaincode.Path = "/dummy"

	if _, err := client.LifecycleInstallChaincode(chaincode); err == nil {
		t.Error("should have returned an error, invalid chaincode path provided")
	}

	if client.IsChaincodeInstalled("dummyChaincodePackageID") {
		t.Errorf("chaincode '%s' should not be installed on all peers belonging to org MSP", chaincode.Name)
	}

	if client.IsChaincodeApproved(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should not be approved on channel %s", chaincode.Name, channel.Name)
	}

	if err := client.LifecycleCommitChaincode(channel.Name, chaincode); err == nil {
		t.Error("should have returned an error, chaincode neither installed nor approved")
	}

	if client.IsChaincodeCommitted(channel.Name, chaincode.Name, 1) {
		t.Errorf("chaincode '%s' should not be committed on channel '%s'", chaincode.Name, channel.Name)
	}
}
