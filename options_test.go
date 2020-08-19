package fabclient

import (
	"testing"
	"time"
)

func TestOptionsWithChaincodeID(t *testing.T) {
	opts := &options{
		chaincodeID: "",
	}

	opt := WithChaincodeID("chaincodeID")

	if opt == nil {
		t.Fail()
	}

	opt.apply(opts)

	if opts.chaincodeID != "chaincodeID" {
		t.Fail()
	}
}

func TestOptionsWithChannelID(t *testing.T) {
	opts := &options{
		channelID: "",
	}

	opt := WithChannelID("channelID")

	if opt == nil {
		t.Fail()
	}

	opt.apply(opts)

	if opts.channelID != "channelID" {
		t.Fail()
	}
}

func TestOptionsWithOrdererResponseTimeout(t *testing.T) {
	opts := &options{
		ordererResponseTimeout: -1,
	}

	opt := WithOrdererResponseTimeout(5 * time.Second)

	if opt == nil {
		t.Fail()
	}

	opt.apply(opts)

	if opts.ordererResponseTimeout != 5*time.Second {
		t.Fail()
	}
}

func TestOptionsWithUserIdentity(t *testing.T) {
	opts := &options{
		userIdentity: "",
	}

	opt := WithUserIdentity("foo")

	if opt == nil {
		t.Fail()
	}

	opt.apply(opts)

	if opts.userIdentity != "foo" {
		t.Fail()
	}
}
