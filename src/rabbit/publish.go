package rabbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"strconv"
)

type TSPublish struct {
	Packet []byte
	Name   string
	Host   string
	Port   int64
	User   string
	Pass   string
	Conn   *amqp.Connection
	Chan   *amqp.Channel
	Que    amqp.Queue
}

func (pub *TSPublish) Init() {
	pub.Connection()
	//defer pub.Conn.Close()
	pub.Channel()
	//defer pub.Chan.Close()
	pub.Queue()
}

func (pub *TSPublish) Connection() {
	var err error
	pub.Conn, err = amqp.Dial("amqp://" + pub.User + ":" + pub.Pass + "@" + pub.Host + ":" + strconv.FormatInt(pub.Port, 10) + "/")
	Fail(err, "Failed to connect to RabbitMQ")
}

func (pub *TSPublish) Channel() {
	var err error
	pub.Chan, err = pub.Conn.Channel()
	Fail(err, "Failed to open a channel")
}

func (pub *TSPublish) Queue() {
	var err error
	fmt.Println(pub.Name)
	pub.Que, err = pub.Chan.QueueDeclare(
		pub.Name, // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	Fail(err, "Failed to open a channel")
}

func (pub *TSPublish) Do(packet []byte) {
	var err error
	if packet == nil {
	} else {
		pub.Packet = make([]byte, len(packet))
		copy(pub.Packet, packet)
	}
	if pub.Packet == nil {

	} else {

		if pub.Chan == nil {
			fmt.Println("Chan is nil")
		} else {

			err = pub.Chan.Publish(
				"",           // exchange
				pub.Que.Name, // routing key
				false,        // mandatory
				false,        // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        pub.Packet,
				})
			Fail(err, "Publish error")
		}
	}

}
