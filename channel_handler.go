package fabclient

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type channelHandler interface {
	invoke(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error)
	query(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error)
}

type channelHandlerClient struct {
	client           *channel.Client
	underlyingLedger *ledger.Client
}

func newChannelHandler(ctx context.ChannelProvider) (channelHandler, error) {
	channelClient, err := channel.New(ctx)
	if err != nil {
		return nil, err
	}

	ledgerClient, err := ledger.New(ctx)
	if err != nil {
		return nil, err
	}

	client := &channelHandlerClient{
		client:           channelClient,
		underlyingLedger: ledgerClient,
	}

	return client, nil
}

func (chn *channelHandlerClient) invoke(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	response, err := chn.client.Execute(convertChaincodeRequest(request), convertOptions(opts...)...)
	if err != nil {
		return nil, err
	}
	return convertChaincodeTransactionResponse(response), nil
}

func (chn *channelHandlerClient) query(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	response, err := chn.client.Query(convertChaincodeRequest(request), convertOptions(opts...)...)
	if err != nil {
		return nil, err
	}
	return convertChaincodeTransactionResponse(response), nil
}

func convertBlock(b *common.Block) *Block {
	header := &BlockHeader{
		Number:       b.GetHeader().Number,
		PreviousHash: b.GetHeader().PreviousHash,
		DataHash:     b.GetHeader().DataHash,
	}

	data := &BlockData{
		Data: b.GetData().Data,
	}

	metadata := &BlockMetadata{
		Metadata: b.GetMetadata().Metadata,
	}

	block := &Block{
		Header:   header,
		Data:     data,
		Metadata: metadata,
	}

	return block
}

func convertChaincodeEvent(e *fab.CCEvent) *ChaincodeEvent {
	event := ChaincodeEvent(*e)
	return &event
}

func convertChaincodeRequest(request *ChaincodeRequest) channel.Request {
	invocationChain := make([]*fab.ChaincodeCall, 0, len(request.InvocationChain))
	for _, invoc := range request.InvocationChain {
		invocationChain = append(invocationChain, &fab.ChaincodeCall{
			ID:          invoc.ID,
			Collections: invoc.Collections,
		})
	}

	return channel.Request{
		Args:            convertArrayOfStringsToArrayOfByteArrays(request.Args),
		Fcn:             request.Function,
		ChaincodeID:     request.ChaincodeID,
		TransientMap:    request.TransientMap,
		InvocationChain: invocationChain,
	}
}

func convertChaincodeTransactionResponse(response channel.Response) *TransactionResponse {
	return &TransactionResponse{
		Payload:       response.Payload,
		Status:        response.ChaincodeStatus,
		TransactionID: string(response.TransactionID),
	}
}

func convertOptions(opts ...Option) []channel.RequestOption {
	convertedOpts := make([]channel.RequestOption, 0, len(opts))

	o := &options{
		ordererResponseTimeout: -1,
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	if o.ordererResponseTimeout != -1 {
		convertedOpts = append(convertedOpts, channel.WithTimeout(fab.OrdererResponse, o.ordererResponseTimeout))
	}

	return convertedOpts
}
