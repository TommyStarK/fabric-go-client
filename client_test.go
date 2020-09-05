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
		t.Log("should have failed: path towards a not supported extension file")
		t.Fail()
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
	testMembershipServiceProvider(t, org1client.msp, org1client.Config())
}

func TestCreateUpdateJoinChannelForOrg1AndOrg2(t *testing.T) {
	org1CreateUpdateAndJoinChannel(t, org1client)
	org2UpdateAndJoinChannel(t, org2client)
}

func TestChannelManagementFailureCases(t *testing.T) {
	channelManagementFailureCases(t, org1client)
}

func TestInstallAndApproveChaincodeOnOrg1AndOrg2(t *testing.T) {
	org1InstallAndApproveChaincodeContractAPI(t, org1client)
	org2InstallAndApproveChaincodeContractAPI(t, org2client)
}

func TestCommitChaincodeForOrg1(t *testing.T) {
	org1CommitChaincode(t, org1client)
}

func TestChaincodeManagementFailureCases(t *testing.T) {
	chaincodeManagementFailureCases(t, org1client)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
