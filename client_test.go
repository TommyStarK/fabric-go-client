package fabclient

import (
	"os"
	"testing"
)

var (
	defaultClient *Client
	defaultConfig *Config
	err           error
)

func TestNewClient(t *testing.T) {
	if _, err := NewClientFromConfigFile("./go.mod"); err == nil {
		t.Log("should have failed: path towards a not supported extension file")
		t.Fail()
	}

	defaultClient, err = NewClientFromConfigFile("./testdata/client/client-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	defaultConfig = defaultClient.Config()
}

func TestMembershipServiceProvider(t *testing.T) {
	testMembershipServiceProvider(t, defaultClient.msp, defaultConfig)
}

func TestChannelResourceManagement(t *testing.T) {
	channel := defaultClient.Config().Channels[0]

	if err := defaultClient.CreateChannel(channel); err != nil {
		t.Fatal(err)
	}

	if err = defaultClient.AnchorPeerSetup(channel); err != nil {
		t.Fatal(err)
	}

	if err = defaultClient.JoinChannel(channel); err != nil {
		t.Fatal(err)
	}
}

func TestChaincodeShimAPI(t *testing.T) {
	chaincode := defaultClient.Config().Chaincodes[0]
	if err := defaultClient.InstallChaincode(chaincode); err != nil {
		t.Fatal(err)
	}
}

func TestChaincodeContractAPILifecycle(t *testing.T) {
	// chaincode := defaultClient.Config().Chaincodes[1]
	// if err := defaultClient.InstallChaincode(chaincode); err != nil {
	// 	t.Fatal(err)
	// }
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
