package fabclient

import (
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
