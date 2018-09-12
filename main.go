package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
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
	istioDirectory     string
}

type OSBClient struct {
	*http.Client
}

var (
	proxyConfig = ProxyConfig{port: DefaultPort, httpClientFactory: httpClientFactory, httpRequestFactory: httpRequestFactory}
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

func noOpHandleRequestCompleted([]byte, []byte) error {
	return nil
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
	handleRequestCompleted := noOpHandleRequestCompleted
	if proxyConfig.providerId != "" {
		serviceId := ctx.Params.ByName("instance_id")
		bindingId := ctx.Params.ByName("binding_id")
		systemDomain := proxyConfig.SystemDomain
		providerId := proxyConfig.providerId
		transformResponse = chainTransform(endpoints.GenerateEndpoint, profiles.AddIstioNetworkDataToResponse(providerId, serviceId, systemDomain, proxyConfig.loadBalancerPort))
		handleRequestCompleted = config.WriteIstioFilesForProvider(proxyConfig.istioDirectory, bindingId)
	} else if proxyConfig.consumerId != "" {
		consumerId := proxyConfig.consumerId
		transformRequest = profiles.AddIstioNetworkDataToRequest(consumerId)
	}

	client.forwardAndTransform(ctx, transformRequest, transformResponse, handleRequestCompleted)
}

func (client OSBClient) forward(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	url := createNewUrl(proxyConfig.forwardURL, request)
	proxyRequest, err := proxyConfig.httpRequestFactory(request.Method, url, request.Body)
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

func (client OSBClient) forwardAndTransform(ctx *gin.Context, transformRequest func([]byte) ([]byte, error), transformResponse func([]byte) ([]byte, error),
	handleRequestCompleted func([]byte, []byte) error) {
	writer := ctx.Writer
	request := ctx.Request

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	// we need to buffer the requestBody if we want to read it here and send it
	// in the request.
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	requestBody, err = transformRequest(requestBody)
	log.Printf("translatedRequestBody:\n %v", string(requestBody))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	proxyRequest := createForwardingRequest(request, err, requestBody)

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
	responseBody, err := ioutil.ReadAll(response.Body)
	log.Printf("respBody:\n %v", string(responseBody))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	okResponse := response.StatusCode/100 == 2
	if okResponse {
		responseBody, err = transformResponse(responseBody)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			log.Printf("ERROR: %s\n", err.Error())
			return
		}
		log.Printf("translatedResponseBody:\n %v", string(responseBody))
		handleRequestCompleted(requestBody, responseBody)
	}
	count, err := writer.Write(responseBody)

	fmt.Printf("count: %d\n", count)
	fmt.Printf("error: %v\n", err)

	//reassign responseBody for dump
	response.Body = ioutil.NopCloser(bytes.NewReader(responseBody))
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
	url := createNewUrl(proxyConfig.forwardURL, request)
	proxyRequest, err := proxyConfig.httpRequestFactory(request.Method, url, bytes.NewReader(body))
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
		proxyConfig.port, err = strconv.Atoi(portAsString)
		if nil != err {
			proxyConfig.port = DefaultPort
		}
	}
}

func main() {
	flag.IntVar(&proxyConfig.port, "port", DefaultPort, "port to be used")
	flag.StringVar(&proxyConfig.forwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.StringVar(&proxyConfig.SystemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&proxyConfig.providerId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")
	flag.StringVar(&proxyConfig.consumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")
	flag.IntVar(&proxyConfig.loadBalancerPort, "loadBalancerPort", 0, "port of the load balancer of the landscape")
	flag.StringVar(&proxyConfig.istioDirectory, "istioDirectory", os.TempDir(), "Directory to store the istio configuration files")
	flag.Parse()
	readPort()

	log.Printf("Running on port %d, forwarding to %s\n", proxyConfig.port, proxyConfig.forwardURL)

	router := setupRouter()
	router.Run(fmt.Sprintf(":%d", proxyConfig.port))
}

func setupRouter() *gin.Engine {
	mux := gin.Default()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := OSBClient{proxyConfig.httpClientFactory(tr)}
	if proxyConfig.providerId != "" {
		mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardAndTransformServiceBinding)
	mux.NoRoute(client.forward)

	return mux
}
