package fabclient

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

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
