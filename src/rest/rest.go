package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	//"profile"
	"strconv"
	//"util"
)

type TSKairos struct {
	group []TSKairosMeasurement
}

type TSKairosMeasurement struct {
	name   string
	metric string
	stamp  int64
	value  float64
}

func (m *TSKairosMeasurement) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name      string  `json:"name"`
		Timestamp int64   `json:"timestamp"`
		Value     float64 `json:"value"`
	}{
		Name:      fmt.Sprintf("%s%s", m.name, m.metric),
		Timestamp: m.stamp,
		Value:     m.value,
	})
}

func write(b *bytes.Buffer, a []byte) {
	n, err := (*b).Write(a)
	if n != len(a) {

	}
	if err != nil {

	}
}

func (kdb *TSKairos) Create(name string, metric string, stamp int64, value float64) {
	kdb.group = append(kdb.group, TSKairosMeasurement{name: name, metric: metric, stamp: stamp, value: value})
}

func (kdb *TSKairos) Add(host string, port int64) {
	var url string = "http://"
	var cmd string = "api/v1/datapoints"
	//var cmd string = "api/put/?details&sync"
	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	mJson, _ := json.Marshal(kdb.group)
	fmt.Println(string(mJson))
	resp, _ := http.Post(url, "application/json", bytes.NewReader(mJson))
	if resp == nil {

	} else {
		defer resp.Body.Close()
	}

	kdb.Reset()
}

func (kdb *TSKairos) Reset() {
	kdb.group = []TSKairosMeasurement{}
}
