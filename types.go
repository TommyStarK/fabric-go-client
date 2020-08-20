package fabclient

type BlockData struct {
	Data [][]byte
}

type BlockHeader struct {
	Number       uint64
	PreviousHash []byte
	DataHash     []byte
}

type BlockMetadata struct {
	Metadata [][]byte
}

type Block struct {
	Header   *BlockHeader
	Data     *BlockData
	Metadata *BlockMetadata
}

type Chaincode struct {
	InitArgs []string `json:"initArgs" yaml:"initArgs"`
	Name     string   `json:"name" yaml:"name"`
	Path     string   `json:"path" yaml:"path"`
	Policy   string   `json:"policy,omitempty" yaml:"policy,omitempty"`
	Version  string   `json:"version" yaml:"version"`
}

type ChaincodeCall struct {
	ID          string
	Collections []string
}

type ChaincodeEvent struct {
	TxID        string
	ChaincodeID string
	EventName   string
	Payload     []byte
	BlockNumber uint64
	SourceURL   string
}

type ChaincodeRequest struct {
	ChaincodeID     string
	Function        string
	Args            []string
	TransientMap    map[string][]byte
	InvocationChain []*ChaincodeCall
}

type Channel struct {
	AnchorPeerConfigPath string `json:"anchorPeerConfigPath,omitempty" yaml:"anchorPeerConfigPath,omitempty"`
	ConfigPath           string `json:"configPath" yaml:"configPath"`
	Name                 string `json:"name" yaml:"name"`
}

type Identity struct {
	Certificate string `json:"certificate" yaml:"certificate"`
	PrivateKey  string `json:"privateKey" yaml:"privateKey"`
	Username    string `json:"username" yaml:"username"`
}

type TransactionResponse struct {
	Payload       []byte
	Status        int32
	TransactionID string
}
