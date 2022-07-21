package rmq

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Almazatun/rabbit-go/web/pkg/configs"
	"github.com/Almazatun/rabbit-go/web/pkg/http/inputs"
	utils "github.com/Almazatun/rabbit-go/web/pkg/utils"
)

type RpcResult[T any] struct {
	Method string
	Result T
}

func PublishMessage(regBody inputs.RpcRequest) (res *RpcResult[string]) {
	// fmt.Println("Service RabbitMQ")

	// Local MQ
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	conn, err := amqp.Dial("amqp://" + config.RABBITMQ_DEFAULT_USER +
		":" + config.RABBITMQ_DEFAULT_PASS + "@" + config.MQ_HOST + ":5672/")

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
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	corrId := utils.RandomString(32)

	out, err := json.Marshal(regBody)

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
			var response RpcResult[string]

			err := json.Unmarshal(d.Body, &response)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			// fmt.Println("UPDATE_MESS_RESPONSE", response)

			res = &response

			break
		}
	}

	return res
}
