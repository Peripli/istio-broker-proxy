package router

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

const (
	DefaultPort = 8080
)

type RouterConfig struct {
	ForwardURL         string
	Port               int
	HttpClientFactory  func(tr *http.Transport) *http.Client
	HttpRequestFactory func(method string, url string, body io.Reader) (*http.Request, error)
}

type osbProxy struct {
	*http.Client
	interceptor ServiceBrokerInterceptor
	config      RouterConfig
}

func (client osbProxy) updateCredentials(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request
	log.Printf("Update credentials request: %v %v", request.Method, request.URL.Path)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Received body: %v\n", string(body))
	response, err := client.interceptor.adaptCredentials(body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(response)
}

func (client osbProxy) forward(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	url := createNewUrl(client.config.ForwardURL, request)
	proxyRequest, err := client.config.HttpRequestFactory(request.Method, url, request.Body)
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

func (client osbProxy) forwardBindRequest(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	// we need to buffer the requestBody if we want to read it here and send it
	// in the request.
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		httpError(writer, err)
		return
	}

	var bindRequest model.BindRequest
	err = json.Unmarshal(requestBody, &bindRequest)
	if err != nil {
		httpError(writer, err)
		return
	}

	bindRequest = *client.interceptor.preBind(bindRequest)

	requestBody, err = json.Marshal(bindRequest)
	log.Printf("translatedRequestBody:\n %v", string(requestBody))
	if err != nil {
		httpError(writer, err)
		return
	}
	proxyRequest := client.createForwardingRequest(request, err, requestBody)

	response, err := client.Do(proxyRequest)
	if err != nil {
		httpError(writer, err)
		return
	}
	log.Printf("Request forwarded: %s\n", response.Status)

	defer func() {
		response.Body.Close()
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	log.Printf("respBody:\n %v", string(responseBody))
	if err != nil {
		httpError(writer, err)
		return
	}

	okResponse := response.StatusCode/100 == 2
	if okResponse {
		var bindResponse model.BindResponse
		err = json.Unmarshal(responseBody, &bindResponse)
		if err != nil {
			httpError(writer, err)
			return
		}
		bindingId := ctx.Params.ByName("binding_id")
		modifiedBindResponse, err := client.interceptor.postBind(bindRequest, bindResponse, bindingId)
		if err != nil {
			httpError(writer, err)
			return
		}
		bindResponse = *modifiedBindResponse

		responseBody, err = json.Marshal(bindResponse)
		log.Printf("translatedResponseBody:\n %v", string(responseBody))
	}

	for name, values := range response.Header {
		switch strings.ToLower(name) {
		case "content-length":
			writer.Header()[name] = []string{fmt.Sprintf("%d", len(responseBody))}
		case "transfer-encoding":
			// just remove it
		default:
			writer.Header()[name] = values
		}
	}

	writer.WriteHeader(response.StatusCode)

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

func httpError(writer gin.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
	log.Printf("ERROR: %s\n", err.Error())
}

func httpClientFactory(tr *http.Transport) *http.Client {
	client := &http.Client{Transport: tr}
	return client
}

func httpRequestFactory(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (client osbProxy) createForwardingRequest(request *http.Request, err error, body []byte) *http.Request {
	url := createNewUrl(client.config.ForwardURL, request)
	proxyRequest, err := client.config.HttpRequestFactory(request.Method, url, bytes.NewReader(body))
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

func SetupRouter(interceptor ServiceBrokerInterceptor, routerConfig RouterConfig) *gin.Engine {
	if routerConfig.HttpClientFactory == nil {
		routerConfig.HttpClientFactory = httpClientFactory
	}
	if routerConfig.HttpRequestFactory == nil {
		routerConfig.HttpRequestFactory = httpRequestFactory
	}
	if routerConfig.Port == 0 {
		routerConfig.Port = DefaultPort
	}

	mux := gin.Default()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := osbProxy{routerConfig.HttpClientFactory(tr), interceptor, routerConfig}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardBindRequest)
	mux.NoRoute(client.forward)

	return mux
}
