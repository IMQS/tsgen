package rabbit

import (
	//"fmt"
	"github.com/streadway/amqp"
)

type TSConsume struct {
	Name   string
	Enable bool

	Acknowledge bool

	Conn *amqp.Connection
	Chan *amqp.Channel

	Delivery <-chan amqp.Delivery
	Kill     chan bool
}

func (con *TSConsume) Init() {
	con.Kill = make(chan bool)
}

func (con *TSConsume) Handle() {
	for del := range con.Delivery {
		if len(del.Body) < 0 {
		}
		if con.Acknowledge {
			del.Ack(false)
		}
	}
}

func (con *TSConsume) Listen() {
	var err error
	con.Delivery, err = con.Chan.Consume(
		con.Name, // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	Fail(err, "Publish error")

	go con.Handle()

	// Wait on consumer to be killed
	<-con.Kill
	//con.Close()

}

func (con *TSConsume) Close() {
	con.Conn.Close()
	con.Chan.Close()
}
