package fabclient

import (
	"os"
	"testing"
)

var (
	org1client *Client
	org2client *Client
	err        error
)

func TestNewClients(t *testing.T) {
	if _, err := NewClientFromConfigFile("./go.mod"); err == nil {
		t.Error("should have returned an error, path towards a not supported extension file")
	}

	org1client, err = NewClientFromConfigFile("./testdata/organizations/org1/client-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	org2client, err = NewClientFromConfigFile("./testdata/organizations/org2/client-config.yaml")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMembershipServiceProvider(t *testing.T) {
	if _, err := org1client.msp.createSigningIdentity("", ""); err == nil {
		t.Error("should have failed to create signing identity: invalid certificate")
	}

	if _, err := org1client.msp.createSigningIdentity(org1client.Config().Identities.Admin.Certificate, ""); err == nil {
		t.Error("should have failed to create signing identity: invalid private key")
	}

	testMembershipServiceProvider(t, org1client.msp, org1client.Config())
}

func TestCreateUpdateJoinChannelForOrg1AndOrg2(t *testing.T) {
	org1CreateUpdateAndJoinChannel(t, org1client)
	org2UpdateAndJoinChannel(t, org2client)

	handler, err := org1client.selectChannelHandler()
	if err != nil || handler == nil {
		t.Fail()
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("channelall"))
	if err != nil || handler == nil {
		t.Error("should have succeed to select channel handler for channel: channelall")
	}

	handler, err = org1client.selectChannelHandler(WithUserContext("User1"))
	if err != nil || handler == nil {
		t.Error("should have succeed to select channel handler for user: User1")
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("channelall"), WithUserContext("User1"))
	if err != nil || handler == nil {
		t.Error("should have succeed to select channel handler for channel: channelall and user: User1")
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("dummy"))
	if err == nil || handler != nil {
		t.Error("should have returned an error when selecting channel handler for channel: dummy")
	}

	handler, err = org1client.selectChannelHandler(WithUserContext("foo"))
	if err == nil || handler != nil {
		t.Error("should have returned an error when selecting channel handler for user: foo")
	}

	handler, err = org1client.selectChannelHandler(WithChannelContext("channelall"), WithUserContext("foo"))
	if err == nil || handler != nil {
		t.Error("should have returned an error when selecting channel handler for user: foo")
	}
}

func TestChannelManagementFailureCases(t *testing.T) {
	channelManagementFailureCases(t, org1client)
}

func TestInstallAndApproveChaincodeOnOrg1AndOrg2(t *testing.T) {
	org1InstallAndApproveChaincodeContractAPI(t, org1client)
	org2InstallAndApproveChaincodeContractAPI(t, org2client)
}

func TestCommitChaincodeOnOrg1(t *testing.T) {
	org1CommitChaincode(t, org1client)
}

func TestChaincodeManagementFailureCases(t *testing.T) {
	chaincodeManagementFailureCases(t, org1client)
}

func TestChaincodeOperations(t *testing.T) {
	initChaincode(t, org1client)
	writeToLedger(t, org1client)
	readFromLedger(t, org2client)
	queryBlock(t, org1client)
	queryBlockByTxID(t, org2client)
	queryInfo(t, org1client)
	queryBlockByHash(t, org2client)
	registerChaincodeEvent(t, org1client)
	chaincodeEventTimeout(t, org1client)
	chaincodePrivateDataCollection(t, org1client, org2client)
	chaincodeOpsFailureCases(t, org1client)
	testConvertBlockchainInfo(t)
	testConvertChaincodeRequest(t)
}

func TestGatewayWrapping(t *testing.T) {
	testWalletCapabilities(t, org1client.Config())
	testGatewayCapabilities(t, org1client.Config())
	cleanAfterGatewayTests(t, org1client.Config())
}

func TestCloseClient(t *testing.T) {
	org1client.Close()
	org2client.Close()
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
