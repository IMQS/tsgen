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

type TSWrite struct {
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

	Gap     profile.TSProfile
	Latency profile.TSProfile

	Seed    int64
	SrcSite rand.Source
}

type TSFilter struct {
	ftype   string `json:"type"`
	tagk    string `json:"tagk"`
	filter  string `json:"filter"`
	groupBy bool   `json:"groupBy"`
}

type TSMetric struct {
	aggregator string `json:"aggregator"`
	name       string `json:"metric"`
	rate       bool   `json:"rate"`
	downsample string `json:"downsample"`
	value      string
	unit       string
	limit      int64             `json:"limit"`
	tags       map[string]string `json:"tags"`
	filters    []TSFilter        `json:"filters"`
}

type TSQuery struct {
	dbase  EDBaseType
	start  time.Time
	end    time.Time
	zone   string
	value  string
	unit   string
	metric []TSMetric
}

type TSRead struct {
	DBase  EDBaseType
	single TSQuery
	batch  []TSQuery
	Post   bool

	Id       int64
	Site     string
	Tags     map[string]string
	Val      int64
	Retry    int64
	CntRetry int64
	Once     bool

	Gap     profile.TSProfile
	Latency profile.TSProfile

	Seed    int64
	SrcSite rand.Source
}

type TSQueryData struct {
	metric        string             `json:"metric"`
	tags          map[string]string  `json:"tags"`
	aggregateTags []string           `json:"aggregateTags"`
	dps           map[string]float64 `json:"dps"`
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

func (wr *TSWrite) Init(id int64, post bool) {
	wr.Id = id
	wr.Seed = 100
	wr.SrcSite = rand.NewSource(wr.Seed)
	wr.Post = post
}

func (rd *TSRead) Init(id int64, post bool) {
	rd.Id = id
	rd.Post = post
}

func (wr *TSWrite) Create(metric string, site string, stamp int64, value float64, tags map[string]string) {
	wr.batch = append(wr.batch, TSDataPoint{wr.DBase, metric, site, 0, stamp, value, tags})
}

func (rd *TSRead) Create(name string, site string, start time.Time, end time.Time, tags map[string]string) {
	var metric TSMetric
	var query TSQuery
	metric.aggregator = "sum"
	metric.name = name + site
	metric.rate = false
	metric.unit = "milliseconds"
	metric.limit = 1000
	metric.value = "1"
	query.start = start
	query.end = end
	query.metric = append(query.metric, TSMetric(metric))
	rd.single = TSQuery(query)
}

func (wr *TSWrite) OpenTSDBSingle() []byte {

	mJson := bytes.NewBuffer(make([]byte, 0))
	mJson.Write([]byte(`{`))
	mJson.Write([]byte(`"metric" : `))
	mJson.Write([]byte(strconv.Quote(wr.single.metric)))
	mJson.Write([]byte(`,`))
	mJson.Write([]byte(`"timestamp" : `))
	mJson.Write([]byte(strconv.FormatInt(wr.single.stamp/int64(time.Millisecond), 10)))
	mJson.Write([]byte(`,`))
	mJson.Write([]byte(`"value" : `))
	mJson.Write([]byte(strconv.FormatFloat(wr.single.value, 'f', -1, 64)))
	mJson.Write([]byte(`, `))
	mJson.Write([]byte(`"tags" : `))
	mJson.Write([]byte(`{`))

	var cnt int = 0
	for idx, value := range wr.single.tags {
		mJson.Write([]byte(strconv.Quote(idx) + " : "))
		mJson.Write([]byte(strconv.Quote(value)))
		if cnt < (len(wr.single.tags) - 1) {
			mJson.Write([]byte(`,`))
		}
		cnt++
	}

	mJson.Write([]byte(`}`))
	mJson.Write([]byte(`}`))

	return mJson.Bytes()

}

func (wr *TSWrite) Add(host string, port int64) {
	var url string = "http://"
	var cmd string
	var fSingle bool = false

	switch wr.DBase {
	case KAIROS:
		cmd = "api/v1/datapoints"
	case NEW:
		cmd = "samples"
	case OPEN:
		cmd = "api/put/?details"
		if len(wr.batch) == 1 {
			wr.single = wr.batch[0]
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
	switch wr.DBase {
	case OPEN:
		if fSingle {
			mJson = wr.OpenTSDBSingle()
		} else {
			mJson, _ = json.Marshal(wr.batch)
		}
	default:
		mJson, _ = json.Marshal(wr.batch)
	}
	//fmt.Println(string(mJson))

	if wr.Post {
		wr.CntRetry = 0
		for {
			wr.Latency.Execute.Start(0)
			resp, err := http.Post(url, "application/json", bytes.NewReader(mJson))
			if err != nil {
			}
			if wr.Response(resp) {
				wr.Latency.Execute.Stop()
				wr.Latency.Append(int64(wr.Latency.Execute.Telapsed.Nanoseconds()))
				break
			} else {
				//wr.Gap.Execute.TimeOut(1e9)
				if wr.CntRetry == wr.Retry {
					fmt.Println("Data point retry failure")
					break
				}
				wr.CntRetry++
			}
		}
	}
	wr.Reset()

}

func (wr *TSWrite) Code() int {
	switch wr.DBase {
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

func (rd *TSRead) Code() int {
	switch rd.DBase {
	case KAIROS:
		return 200
	case NEW:
		return 0
	case OPEN:
		return 200
	default:
		return 0
	}
}

func (wr *TSWrite) Response(resp *http.Response) bool {
	var pass bool = true
	if resp == nil {
		pass = false
	} else {
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			pass = false
		}
		if resp.StatusCode != wr.Code() {
			pass = false
		}

		defer resp.Body.Close()
	}
	return pass
}

func (rd *TSRead) Response(resp *http.Response) bool {
	var pass bool = true
	if resp == nil {
		pass = false
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if data != nil {

		}
		if !rd.Once {
			//fmt.Println(resp)
			fmt.Println(string(data))
			rd.Once = true
		}

		if err != nil {
			pass = false
		}
		if resp.StatusCode != rd.Code() {
			pass = false
		} else {
			//fmt.Print(",")
		}

		defer resp.Body.Close()
	}
	return pass
}

func (wr *TSWrite) Reset() {
	wr.batch = make([]TSDataPoint, 0)
}

func (rd *TSRead) Reset() {
	rd.batch = make([]TSQuery, 0)
}

func (rd *TSRead) OpenTSDB() []byte {
	mJson := bytes.NewBuffer(make([]byte, 0))
	mJson.Write([]byte(`{`))
	mJson.Write([]byte(`"start" : `))
	mJson.Write([]byte(strconv.FormatInt(rd.single.start.UnixNano()/int64(time.Millisecond), 10)))
	mJson.Write([]byte(`, `))
	mJson.Write([]byte(`"end" : `))
	mJson.Write([]byte(strconv.FormatInt(rd.single.end.UnixNano()/int64(time.Millisecond), 10)))
	mJson.Write([]byte(`, `))
	mJson.Write([]byte(`"queries" : `))
	mJson.Write([]byte(`[`))
	mJson.Write([]byte(`{`))
	for _, metric := range rd.single.metric {
		mJson.Write([]byte(`"aggregator" : `))
		mJson.Write([]byte(strconv.Quote(metric.aggregator)))
		mJson.Write([]byte(`, `))
		mJson.Write([]byte(`"metric" : `))
		mJson.Write([]byte(strconv.Quote(metric.name)))
		mJson.Write([]byte(`, `))
		mJson.Write([]byte(`"rate" : `))
		mJson.Write([]byte(strconv.FormatBool(metric.rate)))
		/*
			mJson.Write([]byte(`"tags" : `))
			mJson.Write([]byte(`{`))
			var cnt int = 0
			for idx, value := range rd.Tags {
				mJson.Write([]byte(strconv.Quote(idx) + " : "))
				mJson.Write([]byte(strconv.Quote(value)))
				if cnt < (len(rd.Tags) - 1) {
					mJson.Write([]byte(`, `))
				}
				cnt++
			}
		*/

	}

	mJson.Write([]byte(`}`))
	mJson.Write([]byte(`]`))
	mJson.Write([]byte(`}`))

	return mJson.Bytes()
}

func (rd *TSRead) Kairos() []byte {
	mJson := bytes.NewBuffer(make([]byte, 0))
	mJson.Write([]byte(`{`))
	mJson.Write([]byte(`"start_absolute" : `))
	mJson.Write([]byte(strconv.FormatInt(rd.single.start.UnixNano()/int64(time.Millisecond), 10)))
	mJson.Write([]byte(`, `))
	mJson.Write([]byte(`"end_absolute" : `))
	mJson.Write([]byte(strconv.FormatInt(rd.single.end.UnixNano()/int64(time.Millisecond), 10)))
	mJson.Write([]byte(`, `))
	mJson.Write([]byte(`"metrics" : `))
	mJson.Write([]byte(`[`))
	mJson.Write([]byte(`{`))
	for _, metric := range rd.single.metric {

		mJson.Write([]byte(`"tags" : `))

		mJson.Write([]byte(`{`))
		mJson.Write([]byte(`}`))
		mJson.Write([]byte(`, `))

		/*
			var cnt int = 0
			for idx, value := range metric.tags {
				mJson.Write([]byte(strconv.Quote(idx) + " : "))
				mJson.Write([]byte(strconv.Quote(value)))
				if cnt < (len(rd.Tags) - 1) {
					mJson.Write([]byte(`, `))
				}
				cnt++
			}
			mJson.Write([]byte(`, `))
		*/

		mJson.Write([]byte(`"name" : `))
		mJson.Write([]byte(strconv.Quote(metric.name)))
		mJson.Write([]byte(`, `))
		mJson.Write([]byte(`"limit" : `))
		mJson.Write([]byte(strconv.FormatInt(metric.limit, 10)))
		mJson.Write([]byte(`, `))
		mJson.Write([]byte(`"aggregators" : `))
		mJson.Write([]byte(`[`))
		mJson.Write([]byte(`{`))
		mJson.Write([]byte(`"name" : `))
		mJson.Write([]byte(strconv.Quote(metric.aggregator)))
		mJson.Write([]byte(`, `))
		mJson.Write([]byte(`"sampling" : `))
		mJson.Write([]byte(`{`))
		mJson.Write([]byte(`"value" : `))
		mJson.Write([]byte(strconv.Quote(metric.value)))
		mJson.Write([]byte(`, `))
		mJson.Write([]byte(`"unit" : `))
		mJson.Write([]byte(strconv.Quote(metric.unit)))
		mJson.Write([]byte(`}`))
		mJson.Write([]byte(`}`))
		mJson.Write([]byte(`]`))

	}
	mJson.Write([]byte(`}`))
	mJson.Write([]byte(`]`))
	mJson.Write([]byte(`}`))

	return mJson.Bytes()
}

func (rd *TSRead) Query(host string, port int64) {
	var url string = "http://"
	var cmd string
	var fSingle bool = true

	switch rd.DBase {
	case KAIROS:
		cmd = "api/v1/datapoints/query"
	case NEW:

	case OPEN:
		cmd = "api/query"
	case CITUS:

	default:

	}

	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	var mJson = make([]byte, 0)
	switch rd.DBase {
	case KAIROS:
		if fSingle {
			mJson = rd.Kairos()
		} else {
			mJson, _ = json.Marshal(rd.single)
		}
	case OPEN:
		if fSingle {
			mJson = rd.OpenTSDB()
		} else {
			mJson, _ = json.Marshal(rd.single)
		}
	default:
		mJson, _ = json.Marshal(rd.batch)
	}
	//fmt.Println(string(mJson))

	if rd.Post {
		rd.CntRetry = 0

		for {
			rd.Latency.Execute.Start(0)
			resp, err := http.Post(url, "application/json", bytes.NewReader(mJson))
			if err != nil {
			}
			if rd.Response(resp) {
				rd.Latency.Execute.Stop()
				rd.Latency.Append(int64(rd.Latency.Execute.Telapsed.Nanoseconds()))
				break
			} else {
				//wr.Gap.Execute.TimeOut(1e9)
				if rd.CntRetry == rd.Retry {
					fmt.Println("Query retry failure")
					break
				}
				rd.CntRetry++
			}
		}
	}
	rd.Reset()

}
