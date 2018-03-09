# LPGE

LPGE is Acronym of Little Panda Golang Engin. Little Panda is my wife's nickname.

## Goal

+ Good performace, effect both in dev and runtime
+ Good document

## todo

+ message msg processor, support prtobuf/json (processor should not have any function about route);
+ add module(can configure);
+ client should be a module too.
+ network module should support tcp/udp, ipv4/ipv6;
+ database should support mongodb, mysql;
+ module support address(hex/string);
+ msg route to which is decided by application. you can choose a agent itself. Agent has much definition, such as 
	r/w agent, and application agent;
+ new agent func should be configure to gate, and pass it to server. the Agent should implement agent(in gate interface);
+ no master cluster.
+ every module should has destory handler.