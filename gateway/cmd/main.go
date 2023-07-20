package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/Almazatun/rabbit-go/gateway/pkg/common/helper"
	routes "github.com/Almazatun/rabbit-go/gateway/pkg/http"
	"github.com/Almazatun/rabbit-go/gateway/pkg/rmq"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	loadENVs()
	app := fiber.New()
	rm := rmq.New(
		make(map[string]*amqp.Queue),
		make(map[string]*amqp.Channel),
	)

	c := rm.Connect(
		helper.GetEnvVar("RABBITMQ_USER"),
		helper.GetEnvVar("RABBITMQ_PASS"),
		helper.GetEnvVar("RABBITMQ_HOST"),
	)
	rmqClose, e := rm.Init(helper.GetEnvVar("SERVICE_RT"))

	if e != nil {
		panic(e)
	}

	defer rmqClose.MainCh.Close()
	defer c.Close()

	// Start the RabbitMQ consumer
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Consumer running in other thread
	go rm.Consume(wg)

	//Running in main thread
	routes.PublicRoutes(app, rm)
	log.Fatal(app.Listen(helper.GetEnvVar("PORT")))

	wg.Wait()
}

func loadENVs() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("🔴 Error loading .env variables")
	}
}
