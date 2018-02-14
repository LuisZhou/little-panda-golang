package network

type Processor interface {
	Unmarshal(data []byte) (interface{}, error)
	Marshal(msg interface{}) ([][]byte, error)
}
