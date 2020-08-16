package fabclient

import (
	"testing"
)

func testMembershipServiceProvider(t *testing.T, msp membershipServiceProvider, config *Config) {
	if _, err := msp.createSigningIdentity("", ""); err == nil {
		t.Fail()
	}

	if _, err := msp.createSigningIdentity(config.Identities.Admin.Certificate, ""); err == nil {
		t.Fail()
	}

	if _, err := msp.createSigningIdentity(config.Identities.Admin.Certificate, config.Identities.Admin.PrivateKey); err != nil {
		t.Fail()
	}

	if _, err := msp.getSigningIdentity("foo"); err == nil {
		t.Fail()
	}

}
