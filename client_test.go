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

func TestMembershipServiceProvider(t *testing.T) {
	testMembershipServiceProvider(t, org1client.msp, org1client.Config())
}

func TestChannelResourceManagement(t *testing.T) {
	createUpdateAndJoinChannel(t, org1client)
	channelManagementFailureCases(t, org1client)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
