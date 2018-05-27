# LPGE

LPGE is Acronym of Little Panda Golang Engin. Little Panda is my wife's nickname.

## Goal

+ Good performace, effect both in dev and runtime
+ Good document

## todo
+ need another server in gate which is used to cluster.
+ test ws + protobuf.
+ network module should support tcp/udp, ws, ipv4/ipv6;
+ database should support mongodb, mysql;
+ cluster.
+ add websocket: https://github.com/golang/net, also ref: https://github.com/gorilla/websocket
+ make a global db instance 
+ make model https://stackoverflow.com/questions/18926303/iterate-through-the-fields-of-a-struct-in-go
+ https://stackoverflow.com/questions/23723955/how-can-i-pass-a-slice-as-a-variadic-input
	https://stackoverflow.com/questions/24337145/get-name-of-struct-field-using-reflection
	https://golang.org/ref/spec#Passing_arguments_to_..._parameters
+ map string to model meta. (or use subdir ?)
	https://github.com/asaskevich/govalidator/blob/master/utils.go#L107-L119
	https://gist.github.com/regeda/969a067ff4ed6ffa8ed6	
	https://stackoverflow.com/questions/37510763/go-static-member-variable-such-as-oop-langage