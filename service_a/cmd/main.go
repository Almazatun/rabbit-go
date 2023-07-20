package main

import (
	"fmt"

	"github.com/Almazatun/rabbit-go/service_a/pkg/common/helper"
	rmq "github.com/Almazatun/rabbit-go/service_a/pkg/rmq"
	"github.com/joho/godotenv"
)

func main() {
	loadENVs()
	rm := rmq.New()
	c := rm.Connect(
		helper.GetEnvVar("RABBITMQ_USER"),
		helper.GetEnvVar("RABBITMQ_PASS"),
		helper.GetEnvVar("RABBITMQ_HOST"),
	)

	rmqClose, e := rm.Init(helper.GetEnvVar("SERVICE_RT"))

	if e != nil {
		panic(e)
	}

	rm.Consume()

	defer rmqClose.MainCh.Close()
	defer c.Close()
}

func loadENVs() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("🔴 Error loading .env variables")
	}
}
