package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
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
	tags   map[string]string
}

func (m *TSKairosMeasurement) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name      string            `json:"name"`
		Timestamp int64             `json:"timestamp"`
		Value     float64           `json:"value"`
		Tags      map[string]string `json:"tags"`
	}{
		Name:      fmt.Sprintf("%s%s", m.name, m.metric),
		Timestamp: m.stamp / int64(time.Millisecond),
		Value:     m.value,
		Tags:      m.tags,
	})
}

func write(b *bytes.Buffer, a []byte) {
	n, err := (*b).Write(a)
	if n != len(a) {

	}
	if err != nil {

	}
}

func (kdb *TSKairos) Init() {

}

func (kdb *TSKairos) Create(name string, metric string, stamp int64, value float64, tags map[string]string) {
	kdb.group = append(kdb.group, TSKairosMeasurement{name, metric, stamp, value, tags})
}

func (kdb *TSKairos) Add(host string, port int64) {
	var url string = "http://"
	var cmd string = "api/v1/datapoints"

	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	mJson, _ := json.Marshal(kdb.group)
	//fmt.Println(string(mJson))

	if mJson != nil {

	}

	resp, _ := http.Post(url, "application/json", bytes.NewReader(mJson))

	if resp == nil {
		fmt.Print("No response")
		os.Exit(1)
	} else {
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
			fmt.Printf("%s\n", string(contents))
			os.Exit(1)
		}

		if resp.StatusCode != 204 {
			fmt.Println()
			fmt.Println("Response code: ", resp.StatusCode) //Uh-oh this means our test failed
			fmt.Println()
			os.Exit(1)
		}

		defer resp.Body.Close()
	}

	kdb.Reset()

}

func (kdb *TSKairos) Reset() {
	kdb.group = make([]TSKairosMeasurement, 0)
	//kdb.group = []TSKairosMeasurement{}
}
