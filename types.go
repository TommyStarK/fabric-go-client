package fabclient

// BlockData holds the transactions
type BlockData struct {
	Data [][]byte
}

// BlockHeader is the element of the block which forms the block chain
type BlockHeader struct {
	Number       uint64
	PreviousHash []byte
	DataHash     []byte
}

// BlockMetadata defines metadata of the block
type BlockMetadata struct {
	Metadata [][]byte
}

// Block is finalized block structure to be shared among the orderer and peer
type Block struct {
	Header   *BlockHeader
	Data     *BlockData
	Metadata *BlockMetadata
}

// Chaincode describes info of a chaincode
type Chaincode struct {
	InitArgs []string `json:"initArgs" yaml:"initArgs"`
	Name     string   `json:"name" yaml:"name"`
	Path     string   `json:"path" yaml:"path"`
	Policy   string   `json:"policy,omitempty" yaml:"policy,omitempty"`
	Version  string   `json:"version" yaml:"version"`
}

// ChaincodeCall contains the ID of the chaincode as well as an optional set of private data collections that may be accessed by the chaincode
type ChaincodeCall struct {
	ID          string
	Collections []string
}

// ChaincodeEvent contains the data for a chaincode event
type ChaincodeEvent struct {
	TxID        string
	ChaincodeID string
	EventName   string
	Payload     []byte
	BlockNumber uint64
	SourceURL   string
}

// ChaincodeRequest contains the parameters to query and execute an invocation transaction
type ChaincodeRequest struct {
	ChaincodeID     string
	Function        string
	Args            []string
	TransientMap    map[string][]byte
	InvocationChain []*ChaincodeCall
}

// Channel describes a channel configuration
type Channel struct {
	AnchorPeerConfigPath string `json:"anchorPeerConfigPath,omitempty" yaml:"anchorPeerConfigPath,omitempty"`
	ConfigPath           string `json:"configPath" yaml:"configPath"`
	Name                 string `json:"name" yaml:"name"`
}

// Identity holds crypto material for creating a signing identity
type Identity struct {
	Certificate string `json:"certificate" yaml:"certificate"`
	PrivateKey  string `json:"privateKey" yaml:"privateKey"`
	Username    string `json:"username" yaml:"username"`
}

// TransactionResponse  contains response parameters for query and execute an invocation transaction
type TransactionResponse struct {
	Payload       []byte
	Status        int32
	TransactionID string
}
