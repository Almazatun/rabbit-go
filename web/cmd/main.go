package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	handlers "github.com/Almazatun/rabbit-go/web/pkg/http/handlers"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/rpc", handlers.Rpc).Methods("POST")

	log.Fatal(http.ListenAndServe(":3001", router))
}
