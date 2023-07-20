package inputs

type RpcInput struct {
	RK     string `json:"rk"`
	Msg    string `json:"message"`
	Method string `json:"method"`
}

type RpcRequest struct {
	Method string
	Msg    string
}

type RmqReplyMsg struct {
	Msg    string
	Action *string
	Method string
}
