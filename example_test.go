package fabclient

import (
	"io/ioutil"
	"log"
	"time"
)

func Example() {
	client, err := NewClientFromConfigFile("./testdata/organizations/org1/client-config.yaml")
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
	user := client.Config().Identities.Users[0]

	cert, err := ioutil.ReadFile(user.Certificate)
	if err != nil {
		log.Fatal(err)
	}

	pk, err := ioutil.ReadFile(user.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	id := NewWalletX509Identity("Org1MSP", string(cert), string(pk))

	w, err := NewFileSystemWallet("./testdata/test-wallet")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.Put(user.Username, id); err != nil {
		log.Fatal(err)
	}

	gtw, err := Connect(WithConfigFromFile(client.Config().ConnectionProfile), WithIdentity(w, user.Username))
	if err != nil {
		log.Fatal(err)
	}

	network, err := gtw.GetNetwork(channel.Name)
	if err != nil {
		log.Fatal(err)
	}

	contract := network.GetContract(chaincode.Name)

	if _, err := contract.SubmitTransaction("Store", []string{"gateway-asset-test", `{"content": "this is another content test"}`}); err != nil {
		log.Fatal(err)
	}

	result, err := contract.EvaluateTransaction("Query", []string{"gateway-asset-test"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(result))
}
