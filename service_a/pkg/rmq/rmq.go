package rmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Almazatun/rabbit-go/service_a/pkg/common/helper"
	"github.com/Almazatun/rabbit-go/service_a/pkg/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ReqMes struct {
	Msg    string
	Action *string
}

type RpcRequest struct {
	Method string
	Msg    string
}

type RpcResult[T any] struct {
	Method string
	Msg    T
}

type RMQ struct {
	Connection *amqp.Connection
	mainQueue  *amqp.Queue
	mainCh     *amqp.Channel
}

type RmqClose struct {
	MainCh *amqp.Channel
}

type RmqReplyMsg struct {
	msg    string
	action *string
}

func New() *RMQ {
	return new(RMQ)
}

func (r *RMQ) Init(name string) (rmqClose *RmqClose, err error) {
	ch, err := r.CreateMainChannel()

	if err != nil {
		return nil, err
	}

	err_ := r.DeclareMainQueue(name)

	if err_ != nil {
		return nil, err_
	}

	e := r.CreateMainExchAndBindMainQ(name)

	if e != nil {
		return nil, e
	}

	action := "*"
	out, err := json.Marshal(
		ReqMes{
			Msg:    helper.GetEnvVar("SERVICE_RT"),
			Action: &action,
		},
	)

	if err != nil {
		log.Println("🔴", err)
		return nil, errors.New("Publish message error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"gateway", // exchange
		"gateway", // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(out),
		})

	if err != nil {
		log.Println("🔴", err)

		return nil, err
	}

	return &RmqClose{MainCh: ch}, nil
}

func (r *RMQ) Consume() {
	msgs, err := r.mainCh.Consume(
		r.mainQueue.Name, // c
		"",               // consumer
		false,            // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	if err != nil {
		log.Println("🔴 Consumer error", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("Received a message: %s", d.Body)
			var reqBody RpcRequest
			var resBody []byte

			rpcReqChan := make(chan RpcRequest)
			rpcResChan := make(chan RpcResult[string])

			go r.rpcMethodHandler(rpcReqChan, rpcResChan)

			err := json.Unmarshal(d.Body, &reqBody)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			rpcReqChan <- reqBody

			close(rpcReqChan)

			for res := range rpcResChan {
				out, err := json.Marshal(res)

				if err != nil {
					fmt.Println(err)
					panic(err)
				}
				resBody = out
			}

			err = r.mainCh.Publish(
				d.ReplyTo, // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					CorrelationId: d.CorrelationId,
					ContentType:   "text/plain",
					Body:          []byte(resBody),
				})

			d.Ack(true)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (r *RMQ) DeclareMainQueue(name string) error {
	args := amqp.Table{}
	args["x-message-ttl"] = 60000

	q, err := r.mainCh.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		log.Println("🔴", err)
		return err
	}

	r.mainQueue = &q

	return nil
}

func (r *RMQ) CreateMainExchAndBindMainQ(name string) error {
	err := r.mainCh.ExchangeDeclare(
		name,     // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	if err != nil {
		log.Println("🔴", err)

		return err
	}

	err = r.mainCh.QueueBind(
		r.mainQueue.Name, // queue name
		name,             // routing key
		name,             // exchange
		false,
		nil,
	)

	if err != nil {
		log.Println("🔴", err)

		return err
	}

	return nil
}

func (r *RMQ) CreateMainChannel() (c *amqp.Channel, err error) {
	ch, err := r.Connection.Channel()

	if err != nil {
		log.Println("🔴 CHANNEL", err)
		return nil, err
	}

	r.mainCh = ch

	return ch, nil
}

func (r *RMQ) Connect(usr, pass, host string) *amqp.Connection {
	// Local MQ
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	conn, err := amqp.Dial("amqp://" + usr + ":" + pass + "@" + host + ":5672/")

	if err != nil {
		log.Fatal("🔴 Connection ", err)
	}

	r.Connection = conn

	return conn
}

func (r *RMQ) rpcMethodHandler(
	rpcHandlerRequest <-chan RpcRequest,
	rpcResponse chan RpcResult[string],
) {
	booUseCase := domain.Boo{}
	fooUseCase := domain.Foo{}

	for rpcRequest := range rpcHandlerRequest {
		switch rpcRequest.Method {
		case "boo.create":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequest.Method,
				Msg:    booUseCase.Create(rpcRequest.Msg) + "-->1",
			}
		case "boo.update":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequest.Method,
				Msg:    booUseCase.Update(rpcRequest.Msg) + "-->2",
			}
		case "foo.create":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequest.Method,
				Msg:    fooUseCase.Create(rpcRequest.Msg) + "-->3",
			}
		case "foo.update":
			rpcResponse <- RpcResult[string]{
				Method: rpcRequest.Method,
				Msg:    fooUseCase.Update(rpcRequest.Msg) + "-->4",
			}
		}
	}

	close(rpcResponse)
}
