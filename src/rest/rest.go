package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	//"os"
	"strconv"
	"time"
)

type EDBaseType string

const (
	KAIROS EDBaseType = "KAIROS"
	OPEN   EDBaseType = "OPEN"
	NEW    EDBaseType = "NEW"
	CITUS  EDBaseType = "CITUS"
)

type TSDataPoint struct {
	dbase  EDBaseType
	metric string
	site   string
	dptype int64
	stamp  int64
	value  float64
	tags   map[string]string
}

type TSDBase struct {
	batch []TSDataPoint

	DBase     EDBaseType
	Id        int64
	Site      string
	IdxSeries int
	Tags      map[string]string
	Val       int64

	Seed    int64
	SrcSite rand.Source
}

type TSResource struct {
	Id   string            `json:"id"`
	attr map[string]string `json:"attributes"`
}

func (dp *TSDataPoint) MarshalJSON() ([]byte, error) {

	switch dp.dbase {
	case KAIROS:
		return json.Marshal(&struct {
			Name      string            `json:"name"`
			Timestamp int64             `json:"timestamp"`
			Value     float64           `json:"value"`
			Tags      map[string]string `json:"tags"`
		}{
			Name:      fmt.Sprintf("%s%s", dp.metric, dp.site),
			Timestamp: dp.stamp / int64(time.Millisecond),
			Value:     dp.value,
			Tags:      dp.tags,
		})
	case OPEN:
		return json.Marshal(&struct {
			Metric    string            `json:"metric"`
			Timestamp int64             `json:"timestamp"`
			Value     float64           `json:"value"`
			Tags      map[string]string `json:"tags"`
		}{
			Metric:    fmt.Sprintf("%s%s", dp.metric, dp.site),
			Timestamp: dp.stamp / int64(time.Millisecond),
			Value:     dp.value,
			Tags:      dp.tags,
		})
	case NEW:
		return json.Marshal(&struct {
			Timestamp int64      `json:"timestamp"`
			Resource  TSResource `json:"resource"`
			Name      string     `json:"name"`
			Type      string     `json:"type"`
			Value     float64    `json:"value"`
		}{
			Timestamp: dp.stamp / int64(time.Millisecond),
			Resource:  TSResource{Id: "localhost:chassis:temps", attr: dp.tags},
			Name:      fmt.Sprintf("%s%v-%v", dp.metric, dp.site),
			Type:      "GAUGE",
			Value:     dp.value,
		})
	case CITUS:
		return json.Marshal(&struct {
			Metric    string
			Timestamp int64
			Value     float64
		}{
			Metric:    fmt.Sprintf("%s%s", dp.metric, dp.site),
			Timestamp: dp.stamp / int64(time.Millisecond),
			Value:     dp.value,
		})
	default:
		return make([]byte, 0), nil
	}

}

func write(b *bytes.Buffer, a []byte) {
	n, err := (*b).Write(a)
	if n != len(a) {

	}
	if err != nil {

	}
}

func (db *TSDBase) Init(id int64) {
	db.Id = id
	db.Seed = 100
	db.SrcSite = rand.NewSource(db.Seed)
}

func (db *TSDBase) Create(metric string, site string, stamp int64, value float64, tags map[string]string) {
	db.batch = append(db.batch, TSDataPoint{db.DBase, metric, site, 0, stamp, value, tags})
}

func (db *TSDBase) Response(resp *http.Response, code int) {
	if resp == nil {
		fmt.Print("No response")
		//os.Exit(1)
	} else {
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
			fmt.Printf("%s\n", string(contents))
			//os.Exit(1)
		}

		if resp.StatusCode != code {
			fmt.Println()
			fmt.Println("Response code: ", resp.StatusCode) //Uh-oh this means our test failed
			fmt.Println()
			//os.Exit(1)
		}

		defer resp.Body.Close()
	}
}

func (db *TSDBase) Add(host string, port int64) {
	var url string = "http://"
	var cmd string
	switch db.DBase {
	case KAIROS:
		cmd = "api/v1/datapoints"
	case NEW:
		cmd = "samples"
	case OPEN:
		cmd = "api/put/?details"
	case CITUS:
		cmd = "citus"
	default:

	}

	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	mJson, _ := json.Marshal(db.batch)
	//fmt.Println(string(mJson))

	if mJson != nil {

	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(mJson))

	if err != nil {
		fmt.Printf("%s", err)
		//os.Exit(1)
	}

	switch db.DBase {
	case KAIROS:
		db.Response(resp, 204)
	case NEW:
		db.Response(resp, 201)
	case OPEN:
		db.Response(resp, 200)
	default:
	}

	db.Reset()

}

func (db *TSDBase) Reset() {
	db.batch = make([]TSDataPoint, 0)
}
