package rmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Almazatun/rabbit-go/api/pkg/configs"
	domain "github.com/Almazatun/rabbit-go/api/pkg/domain"
	utils "github.com/Almazatun/rabbit-go/api/pkg/utils"
)

type RpcRequest struct {
	Method string
	Mess   string
}

type RpcResult[T any] struct {
	Method string
	Result T
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
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("Received a message: %s", d.Body)
			var reqBody RpcRequest
			var resBody []byte

			rpcReqChannel := make(chan RpcRequest)
			rpcResChannel := make(chan RpcResult[string])

			go r.rpcMethodHandler(rpcReqChannel, rpcResChannel)

			err := json.Unmarshal(d.Body, &reqBody)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			rpcReqChannel <- reqBody

			close(rpcReqChannel)

			for res := range rpcResChannel {
				out, err := json.Marshal(res)

				if err != nil {
					fmt.Println(err)
					panic(err)
				}
				resBody = out
			}

			// out, err := json.Marshal(reqBody)
			// out, err := json.Marshal(r.updateMessBody(&reqBody))

			// if err != nil {
			// 	fmt.Println(err)
			// 	panic(err)
			// }

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					CorrelationId: d.CorrelationId,
					ContentType:   "text/plain",
					Body:          []byte(resBody),
				})

			d.Ack(false)
		}
	}()

	// fmt.Println("Successfully connected to RabbitMQ")

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (r *RMQ) rpcMethodHandler(
	rpcHandlerRequest <-chan RpcRequest,
	rpcResponse chan RpcResult[string],
) {
	rand := utils.RandomString(6)
	booUseCase := domain.Boo{}
	fooUseCase := domain.Foo{}

	for rpcRequestData := range rpcHandlerRequest {
		switch rpcRequestData.Method {
		case "boo.create":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequestData.Method,
				Result: booUseCase.Create(rpcRequestData.Mess) + rand + "-->1",
			}
		case "boo.update":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequestData.Method,
				Result: booUseCase.Update(rpcRequestData.Mess) + rand + "-->2",
			}
		case "foo.create":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequestData.Method,
				Result: fooUseCase.Create(rpcRequestData.Mess) + rand + "-->3",
			}
		case "foo.update":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequestData.Method,
				Result: fooUseCase.Update(rpcRequestData.Mess) + rand + "-->4",
			}
		}
	}

	close(rpcResponse)
}
