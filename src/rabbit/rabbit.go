package rabbit

import (
	"fmt"
)

type ESubscribe string

const (
	PUBLISH ESubscribe = "PUBLISH"
	CONSUME ESubscribe = "CONSUME"
	BOTH    ESubscribe = "BOTH"
)

type TSQueue struct {
	Name      string
	Subscribe ESubscribe
	Publish   []TSPublish
	Consume   []TSConsume
}

func Fail(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func (que *TSQueue) Init() {
	// Start all of the comsumers defined for this TSRabbit instance

	switch que.Subscribe {
	case PUBLISH:
		que.Publish = append(que.Publish, TSPublish{})
	case CONSUME:
		que.Consume = append(que.Consume, TSConsume{})
	case BOTH:
		que.Publish = append(que.Publish, TSPublish{})
		que.Consume = append(que.Consume, TSConsume{})
	default:
	}

}

func (que *TSQueue) Build(user string, pass string, host string, port int64) {

	for _, pub := range que.Publish {
		pub.Name = que.Name
		fmt.Println(pub.Name)
		pub.User = user
		pub.Pass = pass
		pub.Host = host
		pub.Port = port
		pub.Init()
		fmt.Println("Publish")
	}

	for _, con := range que.Consume {
		con.Name = que.Name
		fmt.Println(con.Name)
		con.User = user
		con.Pass = pass
		con.Host = host
		con.Port = port
		con.Init()
		con.Listen()
		fmt.Println("Listen")
	}
}
