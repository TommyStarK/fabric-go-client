package fabclient

import (
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	if _, err := NewConfigFromFile("./testdata/fixtures/client-config.json"); err != nil {
		t.Errorf("should have succeed, error: %w", err)
	}

	if _, err := NewConfigFromFile("./testdata/organizations/org1/client-config.yaml"); err != nil {
		t.Errorf("should have succeed, error: %w", err)
	}

	if _, err := NewConfigFromFile(""); err == nil {
		t.Error("should have returned an error, invalid path")
	}

	if _, err := NewConfigFromFile("./go.mod"); err == nil {
		t.Error("should have returned an error, path towards a not supported extension file")
	}

	if _, err := NewConfigFromFile("./testdata/fixtures/invalid-client-config.json"); err == nil {
		t.Fail()
	}

	if _, err := NewConfigFromFile("./testdata/fixtures/invalid-client-config.yaml"); err == nil {
		t.Fail()
	}
}
