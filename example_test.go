package fabclient

import (
	"log"
	"time"
)

func Example() {
	client, err := NewClientFromConfigFile("./testdata/client/client-config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	channel := client.Config().Channels[0]
	chaincode := client.Config().Chaincodes[0]

	if err := client.SaveChannel(channel.Name, channel.ConfigPath); err != nil {
		log.Fatal(err)
	}

	if err := client.JoinChannel(channel.Name); err != nil {
		log.Fatal(err)
	}

	if err := client.InstallChaincode(chaincode); err != nil {
		log.Fatal(err)
	}

	if err := client.InstantiateOrUpgradeChaincode(channel.Name, chaincode); err != nil {
		log.Fatal(err)
	}

	storeRequest := &ChaincodeRequest{
		ChaincodeID: chaincode.Name,
		Function:    "store",
		Args:        []string{"asset-test", `{"content": "this is a content test"}`},
	}

	storeResult, err := client.Invoke(storeRequest, WithOrdererResponseTimeout(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("store txID: %s", storeResult.TransactionID)

	queryRequest := &ChaincodeRequest{
		ChaincodeID: chaincode.Name,
		Function:    "query",
		Args:        []string{"asset-test"},
	}

	queryResult, err := client.Query(queryRequest)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("query content: %s", string(queryResult.Payload))
}
