package fabclient

import (
	"encoding/json"
	"testing"
	"time"
)

var (
	txID      string
	blockHash []byte
)

func initChaincode(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
		Function:    "init",
		IsInit:      true,
	}

	if _, err := client.Invoke(req, WithOrdererResponseTimeout(2*time.Second)); err != nil {
		t.Fatal(err)
	}
}

func writeToLedger(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
		Function:    "Store",
		Args:        []string{"asset-test", `{"content": "this is a content test"}`},
	}

	res, err := client.Invoke(req, WithOrdererResponseTimeout(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	txID = res.TransactionID
}

func readFromLedger(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
		Function:    "Query",
		Args:        []string{"asset-test"},
	}

	res, err := client.Query(req)
	if err != nil {
		t.Fatal(err)
	}

	var result = struct {
		Content string `json:"content"`
		TxID    string `json:"txID"`
	}{}

	if err := json.Unmarshal(res.Payload, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.TxID) == 0 {
		t.Error("transaction ID should not be empty")
	}

	if result.Content != "this is a content test" {
		t.Error(`content should be: "this is a content test"`)
	}
}

func queryBlock(t *testing.T, client *Client) {
	if _, err := client.QueryBlock(1); err != nil {
		t.Fatal(err)
	}
}

func queryBlockByTxID(t *testing.T, client *Client) {
	if _, err := client.QueryBlockByTxID(txID); err != nil {
		t.Fatal(err)
	}
}

func queryInfo(t *testing.T, client *Client) {
	info, err := client.QueryInfo()
	if err != nil {
		t.Fatal(err)
	}

	if info.Height != 8 {
		t.Error("Height should equal 8")
	}

	blockHash = info.PreviousBlockHash
}

func queryBlockByHash(t *testing.T, client *Client) {
	if _, err := client.QueryBlockByHash(blockHash); err != nil {
		t.Fatal(err)
	}
}

func registerChaincodeEvent(t *testing.T, client *Client) {
	var (
		done        = make(chan bool)
		eventFilter = "test"
		message     = "this is a message test"
	)

	ch, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, eventFilter)
	if err != nil {
		close(done)
		t.Fatal(err)
	}

	go func() {
		select {
		case event := <-ch:
			if event.EventName != eventFilter {
				done <- false
				return
			}

			if string(event.Payload) != message {
				done <- false
				return
			}

			done <- true
			return
		case <-time.After(5 * time.Second):
			done <- false
			return
		}
	}()

	go func() {
		req := &ChaincodeRequest{
			ChaincodeID: client.Config().Chaincodes[0].Name,
			Function:    "SetEvent",
			Args:        []string{eventFilter, message},
		}

		if _, err := client.Invoke(req); err != nil {
			done <- false
		}

		return
	}()

	success := <-done
	if !success {
		t.Error("should have detected the chaincode event")
	}

	close(done)
	client.UnregisterChaincodeEvent(eventFilter)
}

func chaincodeEventTimeout(t *testing.T, client *Client) {
	chEvent, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, "foo")
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan bool)
	go func() {
		select {
		case <-chEvent:
			ch <- true
			return
		case <-time.After(5 * time.Second):
			ch <- false
			return
		}
	}()

	res := <-ch
	if res {
		t.Error("should have timed out when waiting for chaincode event")
	}

	close(ch)
	client.UnregisterChaincodeEvent("foo")
}

func chaincodePrivateDataCollection(t *testing.T, client1, client2 *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client1.Config().Chaincodes[0].Name,
		Function:    "StorePrivateData",
		TransientMap: map[string][]byte{
			"test": []byte("this is a test"),
		},
	}

	if _, err := client1.Invoke(req, WithOrdererResponseTimeout(2*time.Second)); err != nil {
		t.Fatal(err)
	}

	req = &ChaincodeRequest{
		ChaincodeID: client2.Config().Chaincodes[0].Name,
		Function:    "QueryPrivateData",
		Args:        []string{"test"},
	}

	res, err := client2.Query(req, WithOrdererResponseTimeout(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Error("transaction status should equal 200")
	}

	if len(res.TransactionID) == 0 {
		t.Error("TransactionID should not be empty")
	}

	if string(res.Payload) != "this is a test" {
		t.Error("payload should equal 'this is a test'")
	}
}

func chaincodeOpsFailureCases(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
	}

	if _, err := client.Invoke(req); err == nil {
		t.Errorf("should have returned an error when invoking chaincode %+v", req)
	}

	if _, err := client.Invoke(req, WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when invoking chaincode: invalid channel context (dummy)")
	}

	if _, err := client.Query(req); err == nil {
		t.Errorf("should have returned an error when querying chaincode %+v", req)
	}

	if _, err := client.Query(req, WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when querying chaincode: invalid channel context (dummy)")
	}

	if _, err := client.QueryBlockByTxID("dummy"); err == nil {
		t.Error("should have failed to query block by TxID: invalid id (dummy)")
	}

	if _, err := client.QueryBlockByTxID("dummy", WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when querying block by TxID: invalid channel context (dummy)")
	}

	if _, err := client.QueryBlock(0, WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when querying block by TxID: invalid channel context (dummy)")
	}

	if _, err := client.QueryBlockByHash(nil); err == nil {
		t.Error("should have failed to query block by hash: invalid hash (nil)")
	}

	if _, err := client.QueryBlockByHash(nil, WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when querying block by hash: invalid channel context (dummy)")
	}

	if _, err := client.QueryInfo(WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when querying info: invalid channel context (dummy)")
	}

	if _, err := client.RegisterChaincodeEvent("dummy", "dummy", WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when registering chaincode event: invalid channel context (dummy)")
	}

	if err := client.UnregisterChaincodeEvent("dummy", WithChannelContext("dummy")); err == nil {
		t.Error("should have returned an error when unregistering chaincode event: invalid channel context (dummy)")
	}

	if _, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, "eventFilter"); err != nil {
		t.Fatal(err)
	}

	if _, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, "eventFilter"); err == nil {
		t.Errorf("should have returned an error when registering chaincode event: chaincode %s does not exist", client.Config().Chaincodes[0].Name)
	}

	if err := client.UnregisterChaincodeEvent("eventFilter"); err != nil {
		t.Error(err)
	}
}

func testConvertBlockchainInfo(t *testing.T) {
	if bci := convertBlockchainInfo(nil); bci != nil {
		t.Error("blockchain info should be nil")
	}
}

func testConvertChaincodeRequest(t *testing.T) {
	req := &ChaincodeRequest{
		ChaincodeID: "",
		Function:    "",
		Args:        []string{},
		InvocationChain: []*ChaincodeCall{
			{
				ID:          "test",
				Collections: []string{},
			},
		},
		IsInit: true,
	}

	r := convertChaincodeRequest(req)
	if r.InvocationChain[0].ID != "test" {
		t.Fail()
	}

	r = convertChaincodeRequest(nil)
}
