package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

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
