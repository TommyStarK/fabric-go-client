package fabclient

import (
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	if _, err := NewConfigFromFile("./testdata/client/client-config.json"); err != nil {
		t.Logf("should have succeed: %s", err.Error())
		t.Fail()
	}

	if _, err := NewConfigFromFile("./testdata/client/client-config.yaml"); err != nil {
		t.Logf("should have succeed: %s", err.Error())
		t.Fail()
	}

	if _, err := NewConfigFromFile(""); err == nil {
		t.Log("should have failed, invalid path")
		t.Fail()
	}

	if _, err := NewConfigFromFile("./go.mod"); err == nil {
		t.Log("should have failed: path towards a not supported extension file")
		t.Fail()
	}
}
