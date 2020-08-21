package fabclient

import "time"

type options struct {
	channelID              string
	ordererResponseTimeout time.Duration
	userIdentity           string
}

// Option ...
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithChannelID allows to target a specific channel
func WithChannelID(channelID string) Option {
	return optionFunc(func(o *options) {
		o.channelID = channelID
	})
}

// WithOrdererResponseTimeout allows to specify a timeout when sending a transaction to the ordering service
func WithOrdererResponseTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.ordererResponseTimeout = timeout
	})
}

// WithUserIdentity allows to specify a user context
func WithUserIdentity(username string) Option {
	return optionFunc(func(o *options) {
		o.userIdentity = username
	})
}
