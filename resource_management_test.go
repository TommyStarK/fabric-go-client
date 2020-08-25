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
