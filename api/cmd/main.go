package main

import (
	mq "github.com/Almazatun/rabbit-go/api/pkg/rmq"
)

func main() {
	rmq := mq.RMQ{}
	rmq.Rpc()
}
