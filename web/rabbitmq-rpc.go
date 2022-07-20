package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Mess struct {
	Id      string
	Name    string
	Message string
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func PublishMessage(mess string) string {
	fmt.Println("Service RabbitMQ")
	var res string

	// Local MQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	// conn, err := amqp.Dial("amqp://RABBITMQ_DEFAULT_USER:RABBITMQ_DEFAULT_PASS@<HOST>:5672/")

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Successfully connected to RabbitMQ")

	ch, err := conn.Channel()

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"TestQueue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(q)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	corrId := randomString(32)

	reqBody := &Mess{Id: "bla", Name: "BLA", Message: mess}

	out, err := json.Marshal(reqBody)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(out),
		})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	for d := range msgs {
		if corrId == d.CorrelationId {

			fmt.Printf("Received a message: %s", d.Body)
			var response Mess

			err := json.Unmarshal(d.Body, &response)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			fmt.Println("UPDATE_MESS_RESPONSE", response)

			res = response.Message

			break
		}
	}

	return res
}
