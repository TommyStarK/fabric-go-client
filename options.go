package fabclient

import "time"

type options struct {
	chaincodeID            string
	channelID              string
	ordererResponseTimeout time.Duration
	userIdentity           string
}

type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func WithChaincodeID(chaincodeID string) Option {
	return optionFunc(func(o *options) {
		o.chaincodeID = chaincodeID
	})
}

func WithChannelID(channelID string) Option {
	return optionFunc(func(o *options) {
		o.channelID = channelID
	})
}

func WithOrdererResponseTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.ordererResponseTimeout = timeout
	})
}

func WithUserIdentity(username string) Option {
	return optionFunc(func(o *options) {
		o.userIdentity = username
	})
}
