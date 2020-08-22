package fabclient

import (
	"encoding/json"
	"testing"
	"time"
)

var txID string

func writeToLedger(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
		Function:    "store",
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
		Function:    "query",
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

func queryBlockByTxID(t *testing.T, client *Client) {
	if _, err := client.QueryBlockByTxID(txID); err != nil {
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
			Function:    "setEvent",
			Args:        []string{eventFilter, message},
		}

		if _, err := client.Invoke(req); err != nil {
			done <- false
		}

		return
	}()

	success := <-done
	if !success {
		t.Fail()
	}

	close(done)
	client.UnregisterChaincodeEvent(eventFilter)
}

func chaincodeEventTimeout(t *testing.T, client *Client) {
	_, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, "foo")
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan struct{})
	go func() {
		select {
		case <-time.After(5 * time.Second):
			ch <- struct{}{}
			return
		}
	}()

	<-ch
	close(ch)
	client.UnregisterChaincodeEvent("foo")
}

func chaincodeOpsFailureCases(t *testing.T, client *Client) {
	if _, err := client.Invoke(nil); err == nil {
		t.Fail()
	}

	if _, err := client.Invoke(nil, WithChannelContext("dummy")); err == nil {
		t.Fail()
	}

	if _, err := client.Query(nil); err == nil {
		t.Fail()
	}

	if _, err := client.Query(nil, WithChannelContext("dummy")); err == nil {
		t.Fail()
	}

	if _, err := client.QueryBlockByTxID("dummy"); err == nil {
		t.Fail()
	}

	if _, err := client.QueryBlockByTxID("dummy", WithChannelContext("dummy")); err == nil {
		t.Fail()
	}

	if _, err := client.RegisterChaincodeEvent("dummy", "dummy", WithChannelContext("dummy")); err == nil {
		t.Fail()
	}

	if err := client.UnregisterChaincodeEvent("dummy", WithChannelContext("dummy")); err == nil {
		t.Fail()
	}

	if _, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, "eventFilter"); err != nil {
		t.Fatal(err)
	}

	if _, err := client.RegisterChaincodeEvent(client.Config().Chaincodes[0].Name, "eventFilter"); err == nil {
		t.Fail()
	}

	if err := client.UnregisterChaincodeEvent("eventFilter"); err != nil {
		t.Error(err)
	}
}
