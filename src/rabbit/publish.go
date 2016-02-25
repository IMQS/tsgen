package rabbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type TSPublish struct {
	Packet      []byte
	Name        string
	Enable      bool
	Acknowledge bool

	Conn *amqp.Connection
	Chan *amqp.Channel
}

func (pub *TSPublish) Confirm(ack, nack chan uint64) {
	select {

	case <-ack:

	case tag := <-nack:
		fmt.Println("Nack alert! %d", tag)
	}
}

func (pub *TSPublish) Do(packet []byte, ack, nack chan uint64) {
	//fmt.Printf("%#v\n", pub)

	var err error
	/*
		if packet == nil {
			fmt.Print("0")
		} else {
			pub.Packet = make([]byte, len(packet))
			copy(pub.Packet, packet)
		}
		if pub.Packet == nil {
			fmt.Print("0")
		} else {
			if pub.Chan == nil {
				fmt.Print("0")
			} else {
	*/

	if pub.Enable {
		err = pub.Chan.Publish(
			"",       // exchange
			pub.Name, // routing key
			false,    // mandatory
			false,    // immediate
			amqp.Publishing{
				Timestamp:    time.Now(),
				ContentType:  "text/plain",
				Body:         packet,
				DeliveryMode: amqp.Transient,
				Priority:     0,
			})
		Fail(err, "Publish error")

		if pub.Acknowledge {
			pub.Confirm(ack, nack)
		}
	}
	/*
			}
		}
	*/

}

func (pub *TSPublish) Close() {
	pub.Conn.Close()
	pub.Chan.Close()
}
