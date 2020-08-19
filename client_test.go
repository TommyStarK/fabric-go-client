package fabclient

import (
	"os"
	"testing"
)

var (
	org1client *Client
	err        error
)

func TestNewClient(t *testing.T) {
	if _, err := NewClientFromConfigFile("./go.mod"); err == nil {
		t.Log("should have failed: path towards a not supported extension file")
		t.Fail()
	}

	org1client, err = NewClientFromConfigFile("./testdata/client/client-config.yaml")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSelectChannelHandler(t *testing.T) {
	handler, err := org1client.selectChannelHandler(WithChannelID("channelall"))
	if err != nil {
		t.Fatal(err)
	}

	if handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelID("channel12"), WithUserIdentity("User1"))
	if err != nil {
		t.Fatal(err)
	}

	if handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelID("channel12"), WithUserIdentity("foo"))
	if err == nil {
		t.Fail()
	}

	if handler != nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler()
	if err == nil {
		t.Fail()
	}

	if handler != nil {
		t.Fail()
	}
}

func TestMembershipServiceProvider(t *testing.T) {
	testMembershipServiceProvider(t, org1client.msp, org1client.Config())
}

func TestChannelResourceManagement(t *testing.T) {
	createUpdateAndJoinChannel(t, org1client)
	channelManagementFailureCases(t, org1client)
}

func TestChaincodeShimAPIManagement(t *testing.T) {
	installChaincodeShimAPI(t, org1client)
	chaincodeManagementFailureCases(t, org1client)
}

// func TestChaincodeContractAPILifecycle(t *testing.T) {
// 	chaincode := org1client.Config().Chaincodes[1]
// 	if err := org1client.InstallChaincode(chaincode); err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
