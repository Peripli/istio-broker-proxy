package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

//https://10.11.252.10:9293/cf
//https://broker:VoJniQuzmenuhsowelbahenhukejd755@10.11.252.10:9293/cf/v2/catalog -H "X-Broker-API-Version: 2.13"

const (
	DEFAULT_PORT = "8080"
)

var (
	port  string
	count int
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func helloWorld2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World again")
}

func info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "info: %d", count)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	count = count + 1
	//service-fabrik-broker.cf.dev01.aws.istio.sapcloud.io
	resp, err := http.Get("https://broker:VoJniQuzmenuhsowelbahenhukejd755@10.11.252.10:9293/cf/v2/catalog")
	//resp, err := http.Get("https://broker:VoJniQuzmenuhsowelbahenhukejd755@service-fabrik-broker.cf.dev01.aws.istio.sapcloud.io/cf/v2/catalog")

	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		w.WriteHeader(300)
		return
	}

	defer resp.Body.Close()
	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {

	if port = os.Getenv("PORT"); len(port) == 0 {
		port = DEFAULT_PORT
	}
	http.HandleFunc("/hello", helloWorld2)
	http.HandleFunc("/info", info)
	http.HandleFunc("/v2/catalog", redirect)
	http.HandleFunc("/", helloWorld)
	http.ListenAndServe(":"+port, nil)
}
