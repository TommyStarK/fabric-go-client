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

	chaincodePackageID, err := client.LifecycleInstallChaincode(chaincode)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.LifecycleApproveChaincode(channel.Name, chaincodePackageID, chaincode); err != nil {
		log.Fatal(err)
	}

	if err := client.LifecycleCommitChaincode(channel.Name, chaincode); err != nil {
		log.Fatal(err)
	}

	// Init chaincode
	req := &ChaincodeRequest{
		ChaincodeID: chaincode.Name,
		Function:    "init",
		IsInit:      true,
	}

	if _, err := client.Invoke(req, WithOrdererResponseTimeout(2*time.Second)); err != nil {
		log.Fatal(err)
	}

	// Invoke/Query chaincode using default API
	storeRequest := &ChaincodeRequest{
		ChaincodeID: chaincode.Name,
		Function:    "Store",
		Args:        []string{"asset-test", `{"content": "this is a content test"}`},
	}

	storeResult, err := client.Invoke(storeRequest, WithOrdererResponseTimeout(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("store txID: %s", storeResult.TransactionID)

	queryRequest := &ChaincodeRequest{
		ChaincodeID: chaincode.Name,
		Function:    "Query",
		Args:        []string{"asset-test"},
	}

	queryResult, err := client.Query(queryRequest)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("query content: %s", string(queryResult.Payload))

	// using Gateway
}
