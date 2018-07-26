package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"crypto/tls"
	"io/ioutil"
	"bytes"
)

const (
	DefaultPort = "8080"
)

var (
	port  string
	count int
	log = make([]string, 0)
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World again")
}

func info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "info: %d\n", count)
	for _, line := range log {
		fmt.Fprintf(w, "> %s\n", line)
	}
}


func redirect(w http.ResponseWriter, req *http.Request) {
	count = count + 1
	// we need to buffer the body if we want to read it here and send it
	// in the request.

	reqString := fmt.Sprintf(" OldRequest \n %+v", req)
	log = append(log, reqString)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// you can reassign the body if you need to parse it as multipart
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("https://%s%s","10.11.252.10:9293/cf" , req.RequestURI)
	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	reqString = fmt.Sprintf(" NewRequest \n %+v", proxyReq)
	log = append(log, reqString)

	defer resp.Body.Close()

	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {

	if port = os.Getenv("PORT"); len(port) == 0 {
		port = DefaultPort
	}
	http.HandleFunc("/hello", helloWorld)
	http.HandleFunc("/info", info)
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":"+port, nil)
}
