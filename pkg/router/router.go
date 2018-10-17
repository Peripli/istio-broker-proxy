package router

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
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
	var request model.AdaptCredentialsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.HttpErrorFromError(err))
		return
	}
	if len(request.EndpointMappings) == 0 {
		ctx.JSON(http.StatusBadRequest, model.HttpError{Message: "No endpoint mappings available"})
		return
	}
	response, err := model.Adapt(request.Credentials, request.EndpointMappings)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.HttpErrorFromError(err))
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (client osbProxy) adaptCredentials(credentials model.Credentials, mapping []model.EndpointMapping, instanceId string, bindId string, header http.Header) (*model.BindResponse, error) {

	request := model.AdaptCredentialsRequest{Credentials: credentials, EndpointMappings: mapping}
	requestBody, err := json.Marshal(request)

	if nil != err {
		log.Printf("ERROR: %s\n", err.Error())
		return nil, err
	}

	url := fmt.Sprintf("%s/v2/service_instances/%s/service_bindings/%s/adapt_credentials", client.config.ForwardURL, instanceId, bindId)
	proxyRequest, err := client.config.HttpRequestFactory(http.MethodPost, url, bytes.NewReader(requestBody))
	proxyRequest.Header = header

	response, err := client.Do(proxyRequest)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return nil, err
	}
	log.Printf("Response status from adapt credentials: %s\n", response.Status)

	defer response.Body.Close()

	var bindResponse model.BindResponse
	bodyAsBytes, err := ioutil.ReadAll(response.Body)
	if nil != err {
		log.Printf("ERROR: %s\n", err.Error())
		return nil, err
	}
	err = json.Unmarshal(bodyAsBytes, &bindResponse)

	log.Printf("Response from adapt credentials: %#v\n", bindResponse)
	if nil != err {
		log.Printf("ERROR: %s\n", err.Error())
		return nil, err
	}

	return &bindResponse, nil
}

func (client osbProxy) forward(ctx *gin.Context) {
	client.forwardWithCallback(ctx, func(ctx *gin.Context) error {
		return nil
	})
}

func (client osbProxy) deleteBinding(ctx *gin.Context) {
	client.forwardWithCallback(ctx, func(ctx *gin.Context) error {
		return client.interceptor.postDelete(ctx.Params.ByName("binding_id"))
	})
}

func (client osbProxy) forwardWithCallback(ctx *gin.Context, postCallback func(ctx *gin.Context) error) {
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

	if (response.StatusCode / 100) == 2 {
		err = postCallback(ctx)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadGateway)
			log.Printf("ERROR: %s\n", err.Error())
			return
		}
	}

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
		instanceId := ctx.Params.ByName("instance_id")
		modifiedBindResponse, err := client.interceptor.postBind(bindRequest, bindResponse, bindingId,
			func(credentials model.Credentials, mappings []model.EndpointMapping) (*model.BindResponse, error) {
				return client.adaptCredentials(credentials, mappings, instanceId, bindingId, request.Header)
			})
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
	if interceptor.hasAdaptCredentials() {
		mux.POST("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardBindRequest)
	mux.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.deleteBinding)
	mux.NoRoute(client.forward)

	return mux
}
