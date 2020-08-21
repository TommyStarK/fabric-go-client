package fabclient

// BlockData ...
type BlockData struct {
	Data [][]byte
}

// BlockHeader ...
type BlockHeader struct {
	Number       uint64
	PreviousHash []byte
	DataHash     []byte
}

// BlockMetadata ...
type BlockMetadata struct {
	Metadata [][]byte
}

// Block ...
type Block struct {
	Header   *BlockHeader
	Data     *BlockData
	Metadata *BlockMetadata
}

// Chaincode ...
type Chaincode struct {
	InitArgs []string `json:"initArgs" yaml:"initArgs"`
	Name     string   `json:"name" yaml:"name"`
	Path     string   `json:"path" yaml:"path"`
	Policy   string   `json:"policy,omitempty" yaml:"policy,omitempty"`
	Version  string   `json:"version" yaml:"version"`
}

// ChaincodeCall ...
type ChaincodeCall struct {
	ID          string
	Collections []string
}

// ChaincodeEvent ...
type ChaincodeEvent struct {
	TxID        string
	ChaincodeID string
	EventName   string
	Payload     []byte
	BlockNumber uint64
	SourceURL   string
}

// ChaincodeRequest ...
type ChaincodeRequest struct {
	ChaincodeID     string
	Function        string
	Args            []string
	TransientMap    map[string][]byte
	InvocationChain []*ChaincodeCall
}

// Channel ...
type Channel struct {
	AnchorPeerConfigPath string `json:"anchorPeerConfigPath,omitempty" yaml:"anchorPeerConfigPath,omitempty"`
	ConfigPath           string `json:"configPath" yaml:"configPath"`
	Name                 string `json:"name" yaml:"name"`
}

// Identity ...
type Identity struct {
	Certificate string `json:"certificate" yaml:"certificate"`
	PrivateKey  string `json:"privateKey" yaml:"privateKey"`
	Username    string `json:"username" yaml:"username"`
}

// TransactionResponse ...
type TransactionResponse struct {
	Payload       []byte
	Status        int32
	TransactionID string
}
