package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type TSOpen struct {
	group []Datum
}

type Datum struct {
	Metric    string
	Timestamp int64
	Value     float64
	Tags      map[string]string
}

func (m *Datum) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Metric    string            `json:"metric"`
		Timestamp int64             `json:"timestamp"`
		Value     float64           `json:"value"`
		Tags      map[string]string `json:"tags"`
	}{
		Metric:    fmt.Sprintf("%s", m.Metric),
		Timestamp: m.Timestamp,
		Value:     m.Value,
		Tags:      m.Tags,
	})
}

func Append(d *[]Datum, metric string, timestamp int64, value float64, tags map[string]string) {
	*d = append(*d, Datum{metric, timestamp, value, tags})
}

func (open *TSOpen) Create(name string, stamp int64, value float64, tags map[string]string) {
	open.group = append(open.group, Datum{name, stamp, value, tags})
}

func Put(d *[]Datum) {

	userJson, err := json.Marshal(d)
	//os.Stdout.Write(userJson)
	if err != nil {
		fmt.Printf("Parsing data to json failed: %s", err)
	}

	usersUrl := "http://192.168.4.181:4242/api/put/?details&sync"

	request, err := http.NewRequest("POST", usersUrl, bytes.NewBuffer(userJson)) //Create request with JSON body
	if err != nil {
		fmt.Printf("Http request setup failed: %s", err)
	}

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Printf("Request failed: %s", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("%s", err)
		fmt.Printf("%s\n", string(contents))
		os.Exit(1)
	}
	//fmt.Printf("%s\n", string(contents))

	if res.StatusCode != 200 {
		fmt.Println(res)
		fmt.Printf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
		os.Exit(1)
	}
	//os.Stdout.Write(userJson)
}

func Query() {
	usersUrl := "http://192.168.4.66:4242/api/query?ms=true&start=4h-ago&m=sum:stress1{index=20}" //http://192.168.4.181
	resp, err := http.Get(usersUrl)
	if err != nil {
		fmt.Println("Failed to retrieve data from coindesk api: %s", err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", string(contents))

}
