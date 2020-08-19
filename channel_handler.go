package fabclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

type channelHandler interface {
}

type channelHandlerClient struct {
	client           *channel.Client
	underlyingLedger *ledger.Client
}

func newChannelHandler(ctx context.ChannelProvider) (channelHandler, error) {
	client := &channelHandlerClient{}
	return client, nil
}
