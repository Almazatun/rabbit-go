package handlers

import (
	"encoding/json"
	"net/http"

	rm "github.com/Almazatun/rabbit-go/web/pkg/rmq"
)

type RpcRequest struct {
	Mess string
}

func Rpc(w http.ResponseWriter, r *http.Request) {
	var rpcRequest RpcRequest

	err := json.NewDecoder(r.Body).Decode(&rpcRequest)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := rm.PublishMessage(rpcRequest.Mess)

	json.NewEncoder(w).Encode(res)
}
