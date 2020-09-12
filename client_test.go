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
		t.Error("should have returned an error, path towards a not supported extension file")
	}

	org1client, err = NewClientFromConfigFile("./testdata/client/client-config.yaml")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMembershipServiceProvider(t *testing.T) {
	testMembershipServiceProvider(t, org1client.msp, org1client.Config())
}

func TestChannelResourceManagement(t *testing.T) {
	createUpdateAndJoinChannel(t, org1client)
	channelManagementFailureCases(t, org1client)

	handler, err := org1client.selectChannelHandler()
	if err != nil || handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("channelall"))
	if err != nil || handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithUserContext("User1"))
	if err != nil || handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("channelall"), WithUserContext("User1"))
	if err != nil || handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("dummy"))
	if err == nil || handler != nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithUserContext("foo"))
	if err == nil || handler != nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("channelall"), WithUserContext("foo"))
	if err == nil || handler != nil {
		t.Fail()
	}
}

func TestChaincodeShimAPIManagement(t *testing.T) {
	installChaincodeShimAPI(t, org1client)
	chaincodeManagementFailureCases(t, org1client)
}

func TestChaincodeOperations(t *testing.T) {
	writeToLedger(t, org1client)
	readFromLedger(t, org1client)
	queryBlock(t, org1client)
	queryBlockByTxID(t, org1client)
	queryInfo(t, org1client)
	queryBlockByHash(t, org1client)
	registerChaincodeEvent(t, org1client)
	chaincodeEventTimeout(t, org1client)
	chaincodePrivateDataCollection(t, org1client)
	chaincodeOpsFailureCases(t, org1client)
	testConvertChaincodeRequest(t)
}

func TestCloseClient(t *testing.T) {
	org1client.Close()
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
