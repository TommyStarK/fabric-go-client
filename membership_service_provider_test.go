package fabclient

import (
	"testing"
)

func testMembershipServiceProvider(t *testing.T, msp membershipServiceProvider, config *Config) {
	if _, err := msp.createSigningIdentity("", ""); err == nil {
		t.Error("should have returned an error, neither certificate nor private key provided")
	}

	if _, err := msp.createSigningIdentity(config.Identities.Admin.Certificate, ""); err == nil {
		t.Error("should have returned an error, private key not provided")
	}

	if _, err := msp.createSigningIdentity(config.Identities.Admin.Certificate, config.Identities.Admin.PrivateKey); err != nil {
		t.Errorf("should have succeed to create identity, error: %w", err)
	}

	if _, err := msp.getSigningIdentity("foo"); err == nil {
		t.Fail()
	}

}
