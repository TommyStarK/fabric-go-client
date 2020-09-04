package fabclient

import "time"

type options struct {
	channelID              string
	ordererResponseTimeout time.Duration
	userIdentity           string
}

// Option describes a functional parameter for the client.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithChannelContext allows to target a specific channel.
func WithChannelContext(channelID string) Option {
	return optionFunc(func(o *options) {
		o.channelID = channelID
	})
}

// WithOrdererResponseTimeout allows to specify a timeout for orderer response.
func WithOrdererResponseTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.ordererResponseTimeout = timeout
	})
}

// WithUserContext allows to specify a user context.
func WithUserContext(username string) Option {
	return optionFunc(func(o *options) {
		o.userIdentity = username
	})
}
