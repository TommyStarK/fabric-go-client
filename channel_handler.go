package fabclient

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type channelHandler interface {
	invoke(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error)
	query(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error)
	queryBlockByTxID(txID string) (*Block, error)
	registerChaincodeEvent(chaincodeID, eventFilter string) (<-chan *ChaincodeEvent, error)
	unregisterChaincodeEvent(eventFilter string)
}

type ongoingEvent struct {
	registration fab.Registration
	stopChan     chan chan struct{}
	wrapChan     chan *ChaincodeEvent
}

type channelHandlerClient struct {
	client           *channel.Client
	eventManager     *event.Client
	underlyingLedger *ledger.Client

	chaincodeEvents map[string]*ongoingEvent
	mutex           sync.Mutex
}

func newChannelHandler(ctx context.ChannelProvider) (channelHandler, error) {
	channelClient, err := channel.New(ctx)
	if err != nil {
		return nil, err
	}

	eventManager, err := event.New(ctx, event.WithBlockEvents())
	if err != nil {
		return nil, err
	}

	ledgerClient, err := ledger.New(ctx)
	if err != nil {
		return nil, err
	}

	client := &channelHandlerClient{
		client:           channelClient,
		eventManager:     eventManager,
		underlyingLedger: ledgerClient,
		chaincodeEvents:  make(map[string]*ongoingEvent),
		mutex:            sync.Mutex{},
	}

	return client, nil
}

var _ channelHandler = (*channelHandlerClient)(nil)

func (chn *channelHandlerClient) invoke(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	response, err := chn.client.Execute(convertChaincodeRequest(request), convertOptions(opts...)...)
	return convertChaincodeTransactionResponse(response), err
}

func (chn *channelHandlerClient) query(request *ChaincodeRequest, opts ...Option) (*TransactionResponse, error) {
	response, err := chn.client.Query(convertChaincodeRequest(request), convertOptions(opts...)...)
	return convertChaincodeTransactionResponse(response), err
}

func (chn *channelHandlerClient) queryBlockByTxID(txID string) (*Block, error) {
	block, err := chn.underlyingLedger.QueryBlockByTxID(fab.TransactionID(txID))
	return convertBlock(block), err
}

func (chn *channelHandlerClient) registerChaincodeEvent(chaincodeID, eventFilter string) (<-chan *ChaincodeEvent, error) {
	chn.mutex.Lock()
	defer chn.mutex.Unlock()

	if _, ok := chn.chaincodeEvents[eventFilter]; ok {
		return nil, fmt.Errorf("event filter (%s) already registered", eventFilter)
	}

	registration, ch, err := chn.eventManager.RegisterChaincodeEvent(chaincodeID, eventFilter)
	if err != nil {
		return nil, err
	}

	stopChan := make(chan chan struct{})
	wrapChan := make(chan *ChaincodeEvent)
	chn.chaincodeEvents[eventFilter] = &ongoingEvent{
		registration: registration,
		stopChan:     stopChan,
		wrapChan:     wrapChan,
	}

	go func() {
		for {
			select {
			case event := <-ch:
				wrapChan <- convertChaincodeEvent(event)
			case witness := <-stopChan:
				witness <- struct{}{}
				return
			}
		}
	}()

	return wrapChan, nil
}

func (chn *channelHandlerClient) unregisterChaincodeEvent(eventFilter string) {
	chn.mutex.Lock()
	defer chn.mutex.Unlock()

	if _, ok := chn.chaincodeEvents[eventFilter]; ok {
		witness := make(chan struct{})
		ongoingEvent := chn.chaincodeEvents[eventFilter]
		ongoingEvent.stopChan <- witness
		<-witness
		chn.eventManager.Unregister(ongoingEvent.registration)
		close(ongoingEvent.stopChan)
		close(ongoingEvent.wrapChan)
		delete(chn.chaincodeEvents, eventFilter)
	}

	return
}

func convertBlock(b *common.Block) *Block {
	if b == nil {
		return nil
	}

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
	if request == nil {
		return channel.Request{}
	}

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
