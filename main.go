package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
)

const (
	DefaultPort      = "8080"
	ServiceFabrikURL = "10.11.252.10:9293/cf"
)

type ProxyConfig struct {
	ForwardURL string
	port string
}

var (
	config = ProxyConfig{ServiceFabrikURL, DefaultPort}
	count int
	log   = make([]string, 0)
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "info: %d\n", count)
	for _, line := range log {
		fmt.Fprintf(w, "%s\n\n", line)
	}
}

func parseBody(body []byte) []byte {

	// we can parse json and change fields here.
	return body
}

func createNewUrl(newHost string, req *http.Request) (string) {
	return fmt.Sprintf("https://%s%s", newHost, req.URL.Path)
}

func redirect(w http.ResponseWriter, req *http.Request) {
	count = count + 1

	// we need to buffer the body if we want to read it here and send it
	// in the request.

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create a new url from the raw RequestURI sent by the client
	url := createNewUrl(config.ForwardURL, req)
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

	// reassign the body for the dump
	proxyReq.Body = ioutil.NopCloser(bytes.NewReader(body))
	requestDump, err := httputil.DumpRequest(proxyReq, true)
	if err != nil {
		fmt.Println(err)
	}
	log = append(log, "Request: "+string(requestDump))

	defer func() {
		resp.Body.Close()
	}()

	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parseBody(body)

	w.Write(body)

	// reassign body for dump
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log = append(log, "Response: "+string(responseDump))

}

func main() {
	port := os.Getenv("PORT")
	if len(port) != 0 {
		config.port = port
	}
	http.HandleFunc("/hello", helloWorld)
	http.HandleFunc("/info", info)
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":"+config.port, nil)
}
