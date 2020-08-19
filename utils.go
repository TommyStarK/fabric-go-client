package fabclient

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func convertArrayOfStringsToArrayOfByteArrays(args []string) [][]byte {
	res := make([][]byte, 0, len(args))
	for _, arg := range args {
		res = append(res, []byte(arg))
	}
	return res
}

func newBlock(b *common.Block) *Block {
	header := &BlockHeader{
		Number:       b.GetHeader().Number,
		PreviousHash: make([]byte, len(b.GetHeader().PreviousHash)),
		DataHash:     make([]byte, len(b.GetHeader().DataHash)),
	}

	data := &BlockData{
		Data: make([][]byte, len(b.GetData().Data)),
	}

	metadata := &BlockMetadata{
		Metadata: make([][]byte, len(b.GetMetadata().Metadata)),
	}

	copy(header.PreviousHash, b.GetHeader().PreviousHash)
	copy(header.DataHash, b.GetHeader().DataHash)
	copy(data.Data, b.GetData().Data)
	copy(metadata.Metadata, b.GetMetadata().Metadata)

	block := &Block{
		Header:   header,
		Data:     data,
		Metadata: metadata,
	}

	return block
}

func newChaincodeEvent(e *fab.CCEvent) *ChaincodeEvent {
	event := &ChaincodeEvent{
		TxID:        e.TxID,
		ChaincodeID: e.ChaincodeID,
		EventName:   e.EventName,
		Payload:     make([]byte, len(e.Payload)),
		BlockNumber: e.BlockNumber,
		SourceURL:   e.SourceURL,
	}

	copy(event.Payload, e.Payload)
	return event
}
