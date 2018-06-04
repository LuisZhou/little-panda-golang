package network

// Interface of Agent which agents for network client.
type Agent interface {
	// Run can do any kink job, normally receive msg from connect once this method exist, the conn will close.
	Run()
	// OnClose is a callback when conn close.
	OnClose()
}
