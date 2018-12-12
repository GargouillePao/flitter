namespace msgids {
	public const int PID_LOGIN_REQ = 3754879480;
	public const int PID_LOGIN_ACK = 2129065983;

	public map[] msgCreators = {
		3754879480 : () => &LoginReq{},
		2129065983 : () => &LoginAck{},
	}
}