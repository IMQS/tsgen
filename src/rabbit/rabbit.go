package rabbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"strconv"
)

type ESubscribe string

const (
	PUBLISH ESubscribe = "PUBLISH"
	CONSUME ESubscribe = "CONSUME"
)

type TSQueue struct {
	Name        string
	Subscribe   ESubscribe
	Enable      bool
	Acknowledge bool

	Host string
	Port int64
	User string
	Pass string

	Publish []TSPublish
	Consume []TSConsume

	Ack  chan uint64
	NAck chan uint64
}

func Fail(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func (que *TSQueue) Connection() *amqp.Connection {
	var err error
	var conn *amqp.Connection
	//conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	conn, err = amqp.Dial("amqp://" + que.User + ":" + que.Pass + "@" + que.Host + ":" + strconv.FormatInt(que.Port, 10) + "/")

	Fail(err, "Failed to connect to RabbitMQ")
	return conn
}

func (que *TSQueue) Channel(conn *amqp.Connection) *amqp.Channel {
	var err error
	var ch *amqp.Channel
	ch, err = conn.Channel()

	if que.Acknowledge {
		ch.Confirm(false)
		que.Ack, que.NAck = ch.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))
	}

	Fail(err, "Failed to open a channel")

	_, err = ch.QueueDeclare(
		que.Name, // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	Fail(err, "Failed to declare a queue")
	return ch
}

func (que *TSQueue) Init() {
	// Start all of the comsumers defined for this TSRabbit instance
	var conn = que.Connection()
	switch que.Subscribe {
	case PUBLISH:
		que.Publish = append(
			que.Publish,
			TSPublish{nil, que.Name, que.Enable, que.Acknowledge, conn, que.Channel(conn)})
	case CONSUME:
		con := TSConsume{que.Name, que.Enable, que.Acknowledge, conn, que.Channel(conn), nil, nil}
		con.Init()
		que.Consume = append(que.Consume, con)
		if que.Consume[0].Enable {
			// only create additional go routines when testing
			// in normal operation the listener should stay open
			// nad not just while publishing
			que.Consume[0].Listen()
		}
	default:
		//fmt.Printf("%#v\n", pub)
	}

}
