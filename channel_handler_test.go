package fabclient

import (
	"fmt"
	"testing"
)

func storeNewAssetToLedger(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
		Function:    "store",
		Args:        []string{"asset1", `{"content": "debug"}`},
	}

	res, err := client.Invoke(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("\nresponse: %#v\n\n", res)
	// trxID = res.TransactionID
}

func getAssetFromLedger(t *testing.T, client *Client) {
	req := &ChaincodeRequest{
		ChaincodeID: client.Config().Chaincodes[0].Name,
		Function:    "query",
		Args:        []string{"asset1"},
	}

	res, err := client.Query(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("\nresponse: %#v\n\n", res)
	fmt.Printf("payload: %s\n\n", string(res.Payload))
}
