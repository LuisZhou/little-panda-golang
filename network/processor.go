package network

type Processor interface {
	Unmarshal(cmd uint16, data []byte) (interface{}, error)
	Marshal(cmd uint16, msg interface{}) ([]byte, error)
	Register(cmd uint16, msg interface{}) error
}
