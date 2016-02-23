package rabbit

import (
	"github.com/streadway/amqp"
	"strconv"
)

type TSConsume struct {
	Name     string
	Host     string
	Port     int64
	User     string
	Pass     string
	Conn     *amqp.Connection
	Chan     *amqp.Channel
	Que      amqp.Queue
	Delivery <-chan amqp.Delivery

	Kill chan bool
}

func (con *TSConsume) Init() {
	con.Kill = make(chan bool)

	con.Connection()
	defer con.Conn.Close()
	con.Channel()
	defer con.Chan.Close()
	con.Queue()
}

func (con *TSConsume) Connection() {
	var err error
	con.Conn, err = amqp.Dial("amqp://" + con.User + ":" + con.Pass + "@" + con.Host + ":" + strconv.FormatInt(con.Port, 10) + "/")
	Fail(err, "Failed to connect to RabbitMQ")
}

func (con *TSConsume) Channel() {
	var err error
	con.Chan, err = con.Conn.Channel()
	Fail(err, "Failed to open a channel")
}

func (con *TSConsume) Queue() {
	var err error
	con.Que, err = con.Chan.QueueDeclare(
		con.Name, // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	Fail(err, "Failed to open a channel")
}

func (con *TSConsume) Handle() {
	/*
		for del := range con.Delivery {
			if del == nil {

			}
		}
	*/
}

func (con *TSConsume) Listen() {
	var err error
	con.Delivery, err = con.Chan.Consume(
		con.Que.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	Fail(err, "Publish error")

	go con.Handle()
	// Wait on consumer to be killed
	<-con.Kill
}
