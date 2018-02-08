## Keep in mind.
do not think about remote client/server. the rpc is just local.

## todo:

+ I don't think use the callback is a good idea. For the callback can't be serilized for remote server, or two process
	with no share address space.
	So use this mechanism may cause program not portable. Use the return value is just fine. When use the
	Sync rpc, just block, and wait for the result. When use the async rpc, just continue, the return value
	of rpc tell you the call is ok or fail.

+ Every type can cast to interface{}, so use interface{} as the return value of rpc is just ok, we unify the server
	func to func(args ... interface{}) interface{}

+ Address machanism should do in here?

+ If the client is remote, we need a agent for it.

+ How to close server or client.
