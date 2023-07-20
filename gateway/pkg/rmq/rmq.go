package rmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	inputs "github.com/Almazatun/rabbit-go/gateway/pkg/common"
	"github.com/Almazatun/rabbit-go/gateway/pkg/common/helper"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RMQ struct {
	Connection *amqp.Connection
	QueuesMap  map[string]*amqp.Queue
	mainQueue  *amqp.Queue
	mainCh     *amqp.Channel
	chsMap     map[string]*amqp.Channel
	replayChan chan inputs.RmqReplyMsg
	crlIdMap   map[string]string
}

type RmqClose struct {
	MainCh *amqp.Channel
}

func New(
	quesMap map[string]*amqp.Queue,
	chsMap map[string]*amqp.Channel,
) *RMQ {
	return &RMQ{
		QueuesMap: quesMap,
		chsMap:    chsMap,
		crlIdMap:  make(map[string]string),
	}
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

	return &RmqClose{MainCh: ch}, nil
}

func (r *RMQ) DeclareServiceChan(name string) {
	e := r.CreateChannel(name)

	if e != nil {
		log.Println("🟡 Declare service chan error ", e)
	}
}

func (r *RMQ) SendMsgReply(rk, msg, mt string) (res string, err error) {
	var ch *amqp.Channel
	r.replayChan = make(chan inputs.RmqReplyMsg)

	ch, ok := r.chsMap[rk]

	if !ok {
		log.Println("🟡 Invalid routing key")
		return "", errors.New("Invalid rk")
	}

	if ch.IsClosed() {
		c, err := r.Connection.Channel()

		if !ok {
			return "", err
		}

		r.chsMap[rk], ch = c, c
	}

	out, err := json.Marshal(inputs.RpcRequest{
		Msg:    msg,
		Method: mt,
	})

	if err != nil {
		log.Printf("Rk: %s", rk)
		log.Println("🔴", err)

		return "", errors.New("Publish message error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	corrId := helper.RandomString(32)
	r.crlIdMap[corrId] = corrId
	err = ch.PublishWithContext(ctx,
		rk,    // exchange
		rk,    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       r.mainQueue.Name,
			Body:          []byte(out),
		})

	if err != nil {
		log.Println("🔴", err)

		return "", err
	}

	for replyMsg := range r.replayChan {
		res = replyMsg.Msg
		break
	}

	return res, nil
}

func (r *RMQ) Consume(wg *sync.WaitGroup) {
	defer wg.Done()

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

	for msg := range msgs {
		fmt.Printf("Received a message: %s", msg.Body)
		var response inputs.RmqReplyMsg

		err := json.Unmarshal(msg.Body, &response)

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		if response.Action != nil {
			fmt.Println("📍 RK name", response.Msg)
			r.DeclareServiceChan(response.Msg)
		} else {
			_, ok := r.crlIdMap[msg.CorrelationId]

			if ok {
				r.replayChan <- response
				delete(r.crlIdMap, msg.CorrelationId)
			}
		}

		msg.Ack(true)
	}
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

func (r *RMQ) declareQueue(name string) error {
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

	r.QueuesMap[name] = &q

	return nil
}

func (r *RMQ) CreateMainExchAndBindMainQ(name string) error {
	err := r.mainCh.ExchangeDeclare(
		name,     // name
		"direct", // type
		true,     // durable
		true,     // auto-deleted
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

func (r *RMQ) CreateExch(name string) error {
	ch, ok := r.chsMap[name]

	if !ok {
		log.Printf("Ex: %s", name)
		return errors.New("🟡 Channel was not created")
	}

	err := ch.ExchangeDeclare(
		name,     // name
		"direct", // type
		true,     // durable
		true,     // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	if err != nil {
		log.Println("🔴", err)

		return err
	}

	return nil

}

func (r *RMQ) CreateChannel(name string) error {
	ch, err := r.Connection.Channel()

	if err != nil {
		log.Println("🔴", err)
		return err
	}

	r.chsMap[name] = ch

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
