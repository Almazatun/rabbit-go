package rmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Almazatun/rabbit-go/api/pkg/configs"
	utils "github.com/Almazatun/rabbit-go/api/pkg/utils"
)

type MessBody struct {
	Id      string
	Message string
}

type RMQ struct{}

func (r *RMQ) Rpc() {
	// Local MQ
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	conn, err := amqp.Dial("amqp://" + config.RABBITMQ_DEFAULT_USER +
		":" + config.RABBITMQ_DEFAULT_PASS + "@" + config.MQ_HOST + ":5672/")

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
			var reqBody MessBody

			err := json.Unmarshal(d.Body, &reqBody)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			out, err := json.Marshal(r.updateMessBody(&reqBody))

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

	// fmt.Println("Successfully connected to RabbitMQ")

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (r *RMQ) updateMessBody(messBody *MessBody) *MessBody {
	rand := utils.RandomString(6)

	messBody.Message = messBody.Message + string(rand)

	return messBody
}
