package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func RabbitMQRPC() {
	// Local MQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	// conn, err := amqp.Dial("amqp://RABBITMQ_DEFAULT_USER:RABBITMQ_DEFAULT_PASS@<HOST>:5672/")

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer conn.Close()

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

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // c
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("Received a message: %s", d.Body)
			var reqBody Mess

			err := json.Unmarshal(d.Body, &reqBody)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			rand := randomString(6)

			reqBody.Message = reqBody.Message + string(rand)

			out, err := json.Marshal(reqBody)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(out),
				})

			d.Ack(false)

		}
	}()

	fmt.Println("Successfully connected to RabbitMQ")

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
