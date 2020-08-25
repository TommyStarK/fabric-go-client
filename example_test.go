package fabclient

import (
	"log"
)

func Example() {
	client, err := NewClientFromConfigFile("./testdata/client/client-config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	channel := client.Config().Channels[0]

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err != nil {
		log.Fatal(err)
	}

	if err := client.JoinChannel(channel.Name); err != nil {
		log.Fatal(err)
	}
}
