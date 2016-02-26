package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	//"os"
	"profile"
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
	single TSDataPoint
	batch  []TSDataPoint
	Post   bool

	DBase     EDBaseType
	Id        int64
	Site      string
	IdxSeries int
	Tags      map[string]string
	Val       int64
	Retry     int64
	CntRetry  int64

	Gap profile.TSProfile

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

func (db *TSDBase) Init(id int64, post bool) {
	db.Id = id
	db.Seed = 100
	db.SrcSite = rand.NewSource(db.Seed)
	db.Post = post
}

func (db *TSDBase) Create(metric string, site string, stamp int64, value float64, tags map[string]string) {
	db.batch = append(db.batch, TSDataPoint{db.DBase, metric, site, 0, stamp, value, tags})
}

func (db *TSDBase) OpenTSDBSingle() []byte {

	mJson := bytes.NewBuffer(make([]byte, 0))
	mJson.Write([]byte(`{`))
	mJson.Write([]byte(`"metric" : `))
	mJson.Write([]byte(strconv.Quote(db.single.metric)))
	mJson.Write([]byte(`,`))
	mJson.Write([]byte(`"timestamp" : `))
	mJson.Write([]byte(strconv.FormatInt(db.single.stamp/int64(time.Millisecond), 10)))
	mJson.Write([]byte(`,`))
	mJson.Write([]byte(`"value" : `))
	mJson.Write([]byte(strconv.FormatFloat(db.single.value, 'f', -1, 64)))
	mJson.Write([]byte(`, `))
	mJson.Write([]byte(`"tags" : `))
	mJson.Write([]byte(`{`))

	var cnt int = 0
	for idx, value := range db.single.tags {
		mJson.Write([]byte(strconv.Quote(idx) + " : "))
		mJson.Write([]byte(strconv.Quote(value)))
		if cnt < (len(db.single.tags) - 1) {
			mJson.Write([]byte(`,`))
		}
		cnt++
	}

	mJson.Write([]byte(`}`))
	mJson.Write([]byte(`}`))

	return mJson.Bytes()

}

func (db *TSDBase) Add(host string, port int64) {
	var url string = "http://"
	var cmd string
	var fSingle bool = false

	switch db.DBase {
	case KAIROS:
		cmd = "api/v1/datapoints"
	case NEW:
		cmd = "samples"
	case OPEN:
		cmd = "api/put/?details"
		if len(db.batch) == 1 {
			db.single = db.batch[0]
			fSingle = true
		}
	case CITUS:
		cmd = "citus"

	default:

	}

	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	var mJson = make([]byte, 0)
	switch db.DBase {
	case OPEN:
		if fSingle {
			mJson = db.OpenTSDBSingle()
		} else {
			mJson, _ = json.Marshal(db.batch)
		}
	default:
		mJson, _ = json.Marshal(db.batch)
	}
	//fmt.Println(string(mJson))

	if db.Post {
		db.CntRetry = 0
		for {
			resp, err := http.Post(url, "application/json", bytes.NewReader(mJson))
			if err != nil {
			}
			if db.Response(resp, db.Code()) {
				break
			} else {
				//db.Gap.Execute.TimeOut(1e9)

				if db.CntRetry == db.Retry {
					fmt.Println("Retry failure")
					break
				}
				db.CntRetry++
			}
			time.Sleep(100)
		}
	}
	db.Reset()

}

func (db *TSDBase) Code() int {
	switch db.DBase {
	case KAIROS:
		return 204
	case NEW:
		return 201
	case OPEN:
		return 200
	default:
		return 0
	}
}

func (db *TSDBase) Response(resp *http.Response, code int) bool {
	var pass bool = true
	if resp == nil {
		pass = false
	} else {
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			pass = false
		}
		if resp.StatusCode != code {
			pass = false
		}

		defer resp.Body.Close()
	}
	return pass
}

func (db *TSDBase) Reset() {
	db.batch = make([]TSDataPoint, 0)
}
