package msgids 
import "github.com/golang/protobuf/proto"
const(
	PID_LOGIN_REQ uint32 = 3754879480
	PID_LOGIN_ACK uint32 = 2129065983
)
var MsgCreators map[uint32]func() proto.Message = map[uint32]func() proto.Message {
	3754879480 : func()proto.Message { return &LoginReq{}},
	2129065983 : func()proto.Message { return &LoginAck{}},
}
var MsgNames map[uint32] string = map[uint32] string {
	3754879480 : "PID_LOGIN_REQ",
	2129065983 : "PID_LOGIN_ACK",
}
