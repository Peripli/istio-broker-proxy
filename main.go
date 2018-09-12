package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/credentials"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"io"
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
	forwardURL         string
	port               int
	httpClientFactory  func(tr *http.Transport) *http.Client
	httpRequestFactory func(method string, url string, body io.Reader) (*http.Request, error)
	SystemDomain       string
	providerId         string
	consumerId         string
	loadBalancerPort   int
}

type OSBClient struct {
	*http.Client
}

var (
	config = ProxyConfig{port: DefaultPort, httpClientFactory: httpClientFactory, httpRequestFactory: httpRequestFactory}
)

func (client OSBClient) updateCredentials(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request
	log.Printf("Update credentials request: %v %v", request.Method, request.URL.Path)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Received body: %v\n", string(body))
	response, err := credentials.Update(body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(response)
}

func noOpTransform(data []byte) ([]byte, error) {
	return data, nil
}

func chainTransform(transformList ...func([]byte) ([]byte, error)) func([]byte) ([]byte, error) {
	return func(body []byte) ([]byte, error) {
		var err error
		for _, transform := range transformList {
			body, err = transform(body)
			if err != nil {
				return nil, err
			}
		}
		return body, nil
	}
}

func (client OSBClient) forwardAndTransformServiceBinding(ctx *gin.Context) {
	transformRequest := noOpTransform
	transformResponse := noOpTransform

	if config.providerId != "" {
		serviceId := ctx.Params.ByName("instance_id")
		systemDomain := config.SystemDomain
		providerId := config.providerId
		transformResponse = chainTransform(endpoints.GenerateEndpoint, profiles.AddIstioNetworkDataToResponse(providerId, serviceId, systemDomain, config.loadBalancerPort))
	} else if config.consumerId != "" {
		consumerId := config.consumerId
		transformRequest = profiles.AddIstioNetworkDataToRequest(consumerId)
	}

	client.forwardAndTransform(ctx, transformRequest, transformResponse)
}

func (client OSBClient) forward(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	url := createNewUrl(config.forwardURL, request)
	proxyRequest, err := config.httpRequestFactory(request.Method, url, request.Body)
	proxyRequest.Header = request.Header

	response, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	log.Printf("Request forwarded: %s\n", response.Status)

	defer response.Body.Close()

	for name, values := range response.Header {
		writer.Header()[name] = values
	}

	writer.WriteHeader(response.StatusCode)
	io.Copy(writer, response.Body)
}

func (client OSBClient) forwardAndTransform(ctx *gin.Context, transformRequest func([]byte) ([]byte, error), transformResponse func([]byte) ([]byte, error)) {
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

	body, err = transformRequest(body)
	log.Printf("translatedRequestBody:\n %v", string(body))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	proxyRequest := createForwardingRequest(request, err, body)

	response, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	log.Printf("Request forwarded: %s\n", response.Status)

	defer func() {
		response.Body.Close()
	}()

	for name, values := range response.Header {
		writer.Header()[name] = values
	}

	writer.WriteHeader(response.StatusCode)
	body, err = ioutil.ReadAll(response.Body)
	log.Printf("respBody:\n %v", string(body))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	okResponse := response.StatusCode/100 == 2
	if okResponse {
		body, err = transformResponse(body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			log.Printf("ERROR: %s\n", err.Error())
			return
		}
		log.Printf("translatedResponseBody:\n %v", string(body))
	}
	count, err := writer.Write(body)

	fmt.Printf("count: %d\n", count)
	fmt.Printf("error: %v\n", err)

	//reassign body for dump
	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	responseDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}

	log.Printf("Response:\n%v\n", string(responseDump))
}

func httpClientFactory(tr *http.Transport) *http.Client {
	client := &http.Client{Transport: tr}
	return client
}

func httpRequestFactory(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func createForwardingRequest(request *http.Request, err error, body []byte) *http.Request {
	url := createNewUrl(config.forwardURL, request)
	proxyRequest, err := config.httpRequestFactory(request.Method, url, bytes.NewReader(body))
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
	flag.StringVar(&config.SystemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&config.providerId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")
	flag.StringVar(&config.consumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")
	flag.IntVar(&config.loadBalancerPort, "loadBalancerPort", 0, "port of the load balancer of the landscape")
	flag.Parse()
	readPort()

	log.Printf("Running on port %d, forwarding to %s\n", config.port, config.forwardURL)

	router := setupRouter()
	router.Run(fmt.Sprintf(":%d", config.port))
}

func setupRouter() *gin.Engine {
	mux := gin.Default()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := OSBClient{config.httpClientFactory(tr)}
	if config.providerId != "" {
		mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardAndTransformServiceBinding)
	mux.NoRoute(client.forward)

	return mux
}
