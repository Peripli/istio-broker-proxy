package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/credentials"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

const (
	DefaultPort      = 8080
	ServiceFabrikURL = "10.11.252.10:9293/cf"
)

type ProxyConfig struct {
	ForwardURL string
	port       int
}

var (
	config = ProxyConfig{ServiceFabrikURL, DefaultPort}
)

func updateCredentials(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Received update credentials")
	log.Printf("%v\n", string(body))
	response, err := credentials.Update(body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(response)
}

func translateBody(originalRequest *http.Request, responseBody []byte) ([]byte, error) {
	newBody, err := endpoints.GenerateEndpoint(responseBody)
	return newBody, err
}

func createNewUrl(newHost string, req *http.Request) string {

	url := fmt.Sprintf("https://%s%s", newHost, req.URL.Path)

	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.URL.RawQuery)
	}

	return url
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Received redirect")
	log.Printf("%v\n%v\n", req.URL.Path, string(body))

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
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	log.Println("Redirect OK")

	// reassign the body for the dump
	proxyReq.Body = ioutil.NopCloser(bytes.NewReader(body))
	requestDump, err := httputil.DumpRequest(proxyReq, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
	log.Printf("Request:\n%v\n", string(requestDump))

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
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	body, err = translateBody(req, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	w.Write(body)

	// reassign body for dump
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}

	log.Printf("Response:\n%v\n", string(responseDump))

}

func readPort() {
	portAsString := os.Getenv("PORT")
	if len(portAsString) != 0 {
		var err error
		config.port, err = strconv.Atoi(portAsString)
		if nil != err {
			config.port = DefaultPort
		}
	}
}

func main() {
	flag.IntVar(&config.port, "port", DefaultPort, "port to be used")
	flag.StringVar(&config.ForwardURL, "redirectUrl", ServiceFabrikURL, "url for forwarding incoming requests")
	flag.Parse()
	readPort()

	log.Printf("Running on port %d, redirecting to %s\n", config.port, config.ForwardURL)
	log.Println("Starting...")

	http.HandleFunc("/adapt_credentials", updateCredentials)
	http.HandleFunc("/", redirect)
	http.ListenAndServe(fmt.Sprintf(":%d", config.port), nil)
}
