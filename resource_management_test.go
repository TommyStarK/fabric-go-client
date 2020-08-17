package fabclient

import (
	"bytes"
	"testing"
)

func createUpdateAndJoinChannel(t *testing.T, client *Client) {
	channel := client.Config().Channels[0]

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err != nil {
		t.Fatal(err)
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

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err != nil {
		t.Logf("channel %s already exists, should not have returned an error", channel.Name)
		t.Fail()
	}

	channel.Name = "dummy"
	channel.ConfigPath = "/dummy"

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err == nil {
		t.Log("should have returned a non nil error")
		t.Fail()
	}

	if err := client.JoinChannel(channel.Name); err == nil {
		t.Log("should have returned a non nil error")
		t.Fail()
	}
}

func installChaincodeShimAPI(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	if err := client.InstallChaincode(chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstalled(chaincode.Name) {
		t.Logf("chaincode %s should be installed on all peers belonging to org MSP", chaincode.Name)
		t.Fail()
	}

	channelID := client.Config().Channels[0].Name
	if err := client.InstanciateOrUpgradeChaincode(channelID, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstantiated(channelID, chaincode.Name, chaincode.Version) {
		t.Logf("chaincode %s should be instantiated on all peers belonging to org MSP", chaincode.Name)
		t.Fail()
	}

	chaincode.Version = "2.0"
	if err := client.InstallChaincode(chaincode); err != nil {
		t.Fatal(err)
	}

	if err := client.InstanciateOrUpgradeChaincode(channelID, chaincode); err != nil {
		t.Fatal(err)
	}

	if !client.IsChaincodeInstantiated(channelID, chaincode.Name, chaincode.Version) {
		t.Logf("chaincode %s should have been upgraded on all peers belonging to org MSP", chaincode.Name)
		t.Fail()
	}
}

func chaincodeManagementFailureCases(t *testing.T, client *Client) {
	chaincode := client.Config().Chaincodes[0]
	chaincode.Name = "dummy"
	chaincode.Path = "/dummy"

	if err := client.InstallChaincode(chaincode); err == nil {
		t.Log("should have failed, invalid chaincode path provided")
		t.Fail()
	}

	if client.IsChaincodeInstalled(chaincode.Name) {
		t.Logf("chaincode %s should not be installed on all peers belonging to org MSP", chaincode.Name)
		t.Fail()
	}

	channelID := client.Config().Channels[0].Name
	if err := client.InstanciateOrUpgradeChaincode(channelID, chaincode); err == nil {
		t.Log("should have failed, chaincode not installed")
		t.Fail()
	}

	if client.IsChaincodeInstantiated(channelID, chaincode.Name, chaincode.Version) {
		t.Logf("chaincode %s should not be instantiated on all peers belonging to org MSP", chaincode.Name)
		t.Fail()
	}
}

func TestConvertChaincodeInitArgs(t *testing.T) {
	witness := [][]byte{
		[]byte("init"),
		[]byte("a"),
		[]byte("b"),
	}

	test := convertChaincodeInitArgs([]string{"init", "a", "b"})

	if len(witness) != len(test) {
		t.Fail()
	}

	for i := range witness {
		if bytes.Compare(witness[i], test[i]) != 0 {
			t.Fatalf("should be %+v but got %+v", witness[i], test[i])
		}
	}
}
