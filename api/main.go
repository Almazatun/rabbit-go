package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	go RabbitMQRPC()

	router := mux.NewRouter()

	log.Fatal(http.ListenAndServe(":3002", router))

}
