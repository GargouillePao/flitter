package msgids

const PID_LOGIN_ACK uint32 = 2129065983
const PID_LOGIN_REQ uint32 = 3754879480

var MsgCreator map[uint32]func()interface{} = map[uint32]func()interface{}{
	3754879480 : func()interface{} { return &LoginReq{}},
	2129065983 : func()interface{} { return &LoginAck{}},
}