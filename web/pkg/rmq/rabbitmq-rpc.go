package rmq

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Almazatun/rabbit-go/web/pkg/configs"
	utils "github.com/Almazatun/rabbit-go/web/pkg/utils"
)

type MessBody struct {
	Id      string
	Message string
}

func PublishMessage(mess string) (res *MessBody) {
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

	corrId := utils.RandomString(32)

	reqBody := &MessBody{Id: utils.RandomString(5), Message: mess}

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
			var response MessBody

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
