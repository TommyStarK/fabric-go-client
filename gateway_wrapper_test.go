package fabclient

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

var (
	testWallet  *Wallet
	testGateway *Gateway
)

func testWalletCapabilities(t *testing.T, config *Config) {
	user := config.Identities.Users[0]

	cert, err := ioutil.ReadFile(user.Certificate)
	if err != nil {
		t.Fatal(err)
	}

	pk, err := ioutil.ReadFile(user.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	id := NewWalletX509Identity("Org1MSP", string(cert), string(pk))

	if len(id.Certificate()) == 0 {
		t.Fail()
	}

	if len(id.Key()) == 0 {
		t.Fail()
	}

	w, err := NewFileSystemWallet("./testdata/test-wallet")
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Put(user.Username, id); err != nil {
		t.Fatal(err)
	}

	idd, err := w.Get(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	if idd == nil {
		t.Fail()
	}

	if !w.Exists(user.Username) {
		t.Fail()
	}

	list, err := w.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 1 || list[0] != user.Username {
		t.Fail()
	}

	testWallet = w
}

func testGatewayCapabilities(t *testing.T, config *Config) {
	channel := config.Channels[0]
	chaincode := config.Chaincodes[0]
	user := config.Identities.Users[0]

	gtw, err := Connect(WithConfigFromFile(config.ConnectionProfile), WithIdentity(testWallet, user.Username))
	if err != nil {
		t.Fatal(err)
	}

	if gtw == nil {
		t.Fail()
	}

	testGateway = gtw
	network, err := testGateway.GetNetwork(channel.Name)
	if err != nil {
		t.Fatal(err)
	}

	if network.Name() != channel.Name {
		t.Fail()
	}

	contract := network.GetContract(chaincode.Name)

	if contract.Name() != chaincode.Name {
		t.Fail()
	}

	result, err := contract.EvaluateTransaction("Query", []string{"asset-test"})
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fail()
	}

	var asset = struct {
		Content string `json:"content"`
		TxID    string `json:"txID"`
	}{}

	if err := json.Unmarshal(result, &asset); err != nil {
		t.Fatal(err)
	}

	if len(asset.TxID) == 0 {
		t.Error("transaction ID should not be empty")
	}

	if asset.Content != "this is a content test" {
		t.Error(`content should be: "this is a content test"`)
	}

	if _, err := contract.SubmitTransaction("Store", []string{"gateway-asset-test", `{"content": "this is another content test"}`}); err != nil {
		t.Fatal(err)
	}

	result, err = contract.EvaluateTransaction("Query", []string{"gateway-asset-test"})
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fail()
	}

	asset = struct {
		Content string `json:"content"`
		TxID    string `json:"txID"`
	}{}

	if err := json.Unmarshal(result, &asset); err != nil {
		t.Fatal(err)
	}

	if len(asset.TxID) == 0 {
		t.Error("transaction ID should not be empty")
	}

	if asset.Content != "this is another content test" {
		t.Error(`content should be: "this is another content test"`)
	}
}

func cleanAfterGatewayTests(t *testing.T, config *Config) {
	if err := testWallet.Remove(config.Identities.Users[0].Username); err != nil {
		t.Fatal(err)
	}

	testGateway.Close()
}
