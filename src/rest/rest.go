package rest

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"net/http"
	//"profile"
	"strconv"
	"time"

	//"util"
)

type TSKairos struct {
	Json bytes.Buffer
}

type TSNewts struct {
	Json bytes.Buffer
}

func write(b *bytes.Buffer, a []byte) {
	n, err := (*b).Write(a)
	if n != len(a) {

	}
	if err != nil {

	}
}

func (r *TSKairos) Begin() {
	write(&r.Json, []byte(`[`))
}

func (r *TSKairos) Create(name string, stamp int64, value float64) []byte {

	// Clear the buffer
	//r.Reset()

	write(&r.Json, []byte(`{`))

	write(&r.Json, []byte(strconv.Quote(`name`)))
	write(&r.Json, []byte(`:`))
	write(&r.Json, []byte(strconv.Quote(name)))
	write(&r.Json, []byte(`,`))

	write(&r.Json, []byte(strconv.Quote(`timestamp`)))
	write(&r.Json, []byte(`:`))
	write(&r.Json, []byte(strconv.FormatInt(stamp/int64(time.Millisecond), 10)))
	write(&r.Json, []byte(`,`))

	write(&r.Json, []byte(strconv.Quote(`value`)))
	write(&r.Json, []byte(`:`))
	write(&r.Json, []byte(strconv.FormatFloat(value, 'f', -1, 64)))

	write(&r.Json, []byte(`}`))

	return r.Json.Bytes()

}

func (r *TSKairos) Another() {
	write(&r.Json, []byte(`,`))
}

func (r *TSKairos) End() {
	write(&r.Json, []byte(`]`))
}

func (r *TSKairos) Batch(name string, stamp *[]int64, value *[]float64) []byte {

	// Clear the buffer
	r.Reset()

	write(&r.Json, []byte(`[{`))

	write(&r.Json, []byte(strconv.Quote(`name`)))
	write(&r.Json, []byte(`:`))
	write(&r.Json, []byte(strconv.Quote(name)))
	write(&r.Json, []byte(`,`))

	write(&r.Json, []byte(strconv.Quote(`datapoints`)))
	write(&r.Json, []byte(`:`))
	write(&r.Json, []byte(`[`))
	for idx := 0; idx < len(*stamp); idx++ {
		write(&r.Json, []byte(`[`))
		write(&r.Json, []byte(strconv.FormatInt((*stamp)[idx]/int64(time.Millisecond), 10)))
		write(&r.Json, []byte(`,`))
		write(&r.Json, []byte(strconv.FormatFloat((*value)[idx], 'f', -1, 64)))
		write(&r.Json, []byte(`]`))
		if idx < (len(*stamp) - 1) {
			write(&r.Json, []byte(`, `))
		}
	}
	write(&r.Json, []byte(`]`))
	write(&r.Json, []byte(`}]`))

	fmt.Println(string(r.Json.Bytes()))

	return r.Json.Bytes()

}

func (r *TSKairos) Add(host string, port int64) {
	var url string = "http://"
	var cmd string = "api/v1/datapoints"
	//var cmd string = "api/put/?details&sync"
	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	req, _ := http.NewRequest("POST", url, &r.Json)
	if false {
		fmt.Println(req)
	}

	if req != nil {

	}

	fmt.Print("_")

	/*
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			//	panic(err)
		}
		if resp != nil {
			defer resp.Body.Close()
		}
	*/
}

func (r *TSKairos) Reset() {
	r.Json.Reset()
}
