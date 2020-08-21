package fabclient

import (
	"testing"
)

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

func installChaincodeShimAPI(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	if err := client.InstallChaincode(chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstalled(chaincode.Name) {
		t.Errorf("chaincode '%s' should be installed on all peers belonging to org MSP", chaincode.Name)
	}

	channelID := client.Config().Channels[0].Name
	if err := client.InstantiateOrUpgradeChaincode(channelID, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstantiated(channelID, chaincode.Name, chaincode.Version) {
		t.Errorf("chaincode '%s' should be instantiated on all peers belonging to org MSP", chaincode.Name)
	}

	chaincode.Version = "2.0"
	if err := client.InstallChaincode(chaincode); err != nil {
		t.Fatal(err)
	}

	if err := client.InstantiateOrUpgradeChaincode(channelID, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstantiated(channelID, chaincode.Name, chaincode.Version) {
		t.Errorf("chaincode '%s' should have been upgraded on all peers belonging to org MSP", chaincode.Name)
	}
}

func chaincodeManagementFailureCases(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	chaincode.Name = "dummy"
	chaincode.Path = "/dummy"

	if err := client.InstallChaincode(chaincode); err == nil {
		t.Error("should have returned an error, invalid chaincode path provided")
	}

	if client.IsChaincodeInstalled(chaincode.Name) {
		t.Errorf("chaincode '%s' should not be installed on all peers belonging to org MSP", chaincode.Name)
	}

	channelID := client.Config().Channels[0].Name
	if err := client.InstantiateOrUpgradeChaincode(channelID, chaincode); err == nil {
		t.Error("should have returned an error, chaincode not installed")
	}

	if client.IsChaincodeInstantiated(channelID, chaincode.Name, chaincode.Version) {
		t.Errorf("chaincode '%s' should not be instantiated on all peers belonging to org MSP", chaincode.Name)
	}
}
