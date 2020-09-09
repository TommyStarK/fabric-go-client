package fabclient

import (
	"testing"
	"time"
)

func TestOptionsWithChannelContext(t *testing.T) {
	opts := &options{
		channelID: "",
	}

	opt := WithChannelContext("channelID")

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

func TestOptionsWithUserContext(t *testing.T) {
	opts := &options{
		userIdentity: "",
	}

	opt := WithUserContext("foo")

	if opt == nil {
		t.Fail()
	}

	opt.apply(opts)

	if opts.userIdentity != "foo" {
		t.Fail()
	}
}
