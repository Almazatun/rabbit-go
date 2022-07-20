package main

import (
	"encoding/json"
	"net/http"
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

	res := PublishMessage(rpcRequest.Mess)

	json.NewEncoder(w).Encode(res)
}
