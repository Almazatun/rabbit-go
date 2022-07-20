package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/rpc", Rpc).Methods("POST")

	log.Fatal(http.ListenAndServe(":3001", router))
}
