package inputs

type RpcRequest struct {
	Method string `json:"method" validate:"required"`
	Mess   string `json:"message" validate:"required"`
}
