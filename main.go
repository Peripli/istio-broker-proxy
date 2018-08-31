package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
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
	DefaultPort = 8080
)

type ProxyConfig struct {
	forwardURL string
	port       int
}

var (
	config = ProxyConfig{port: DefaultPort}
)

func updateCredentials(ctx *gin.Context) {
	log.Println("Received update credentials")
	writer := ctx.Writer
	request := ctx.Request

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("%v\n", string(body))
	response, err := credentials.Update(body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(response)
}

func translateResponseBody(originalRequest *http.Request, responseBody []byte) ([]byte, error) {
	newBody, err := endpoints.GenerateEndpoint(responseBody)
	return newBody, err
}

func forward(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request
	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	proxyReq := createForwardingRequest(request, err, body)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	log.Printf("Request forwarded: %d %s\n", resp.StatusCode, resp.Status)

	defer func() {
		resp.Body.Close()
	}()

	for name, values := range resp.Header {
		writer.Header()[name] = values
	}

	writer.WriteHeader(resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("respBody:\n %v", string(body))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	log.Printf("translatedBody:\n %v", string(body))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	log.Printf("before write body")

	writer.Write(body)

	// reassign body for dump
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}

	log.Printf("Response:\n%v\n", string(responseDump))

}

func createForwardingRequest(request *http.Request, err error, body []byte) *http.Request {
	url := createNewUrl(config.forwardURL, request)
	proxyRequest, err := http.NewRequest(request.Method, url, bytes.NewReader(body))
	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyRequest.Header = request.Header
	proxyRequest.Header = make(http.Header)
	for key, value := range request.Header {
		proxyRequest.Header[key] = value
	}

	requestDump, err := httputil.DumpRequest(proxyRequest, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
	log.Printf("Proxy request:\n%v\n", string(requestDump))

	return proxyRequest
}

func createNewUrl(newBaseUrl string, req *http.Request) string {
	url := fmt.Sprintf("%s%s", newBaseUrl, req.URL.Path)

	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.URL.RawQuery)
	}

	return url
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
	flag.StringVar(&config.forwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.Parse()
	readPort()

	log.Printf("Running on port %d, forwarding to %s\n", config.port, config.forwardURL)

	router := setupRouter()
	router.Run(fmt.Sprintf(":%d", config.port))
}

func setupRouter() *gin.Engine {

	mux := gin.Default()
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", updateCredentials)
	mux.NoRoute(forward)

	return mux
}
