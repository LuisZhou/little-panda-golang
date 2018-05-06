package network

// Interface of Agent which agents for network client.
type Agent interface {
	// Run do what ever job one agent need to do, once this method exist, the conn will close.
	Run()
	// OnClose is a callback when conn close.
	OnClose()
}
