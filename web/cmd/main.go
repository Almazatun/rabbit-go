package main

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/gorilla/mux"

	handlers "github.com/Almazatun/rabbit-go/web/pkg/http/handlers"
)

var validate *validator.Validate

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/rpc", handlers.Rpc).Methods("POST")

	log.Fatal(http.ListenAndServe(":3001", router))
}
