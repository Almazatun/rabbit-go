package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	inputs "github.com/Almazatun/rabbit-go/web/pkg/http/inputs"
	rm "github.com/Almazatun/rabbit-go/web/pkg/rmq"
	utils "github.com/Almazatun/rabbit-go/web/pkg/utils"
)

func Rpc(w http.ResponseWriter, r *http.Request) {
	var rpcBodyRequest inputs.RpcRequest

	err := json.NewDecoder(r.Body).Decode(&rpcBodyRequest)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e := utils.ValidateRpcRequest[inputs.RpcRequest](rpcBodyRequest)

	if e != nil {
		fmt.Println("------>Validation error", e)
		http.Error(w, "Failed to validate struct", 400)
		return
	}

	if !utils.ValidateRpcMethod(rpcBodyRequest.Method) {
		http.Error(w, "Invalid method", 400)
		return
	}

	res := rm.PublishMessage(rpcBodyRequest)

	json.NewEncoder(w).Encode(res)
}
