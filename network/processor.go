package network

// Interface of processor which do Marshal/Unmarshal.
type Processor interface {
	// Marshal Unmarshal interprete binary data to some type instance accroding the cmd.
	Unmarshal(cmd uint16, data []byte) (interface{}, error)
	// Marshal process msg to binary.
	Marshal(cmd uint16, msg interface{}) ([]byte, error)
	// Register tell processor use what type to interprete data for given cmd.
	Register(cmd uint16, msg interface{}) error
}
