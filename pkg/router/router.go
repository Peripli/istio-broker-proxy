package router

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/credentials"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"io"
	"io/ioutil"
	istioModel "istio.io/istio/pilot/pkg/model"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"strings"
)

const (
	DefaultPort = 8080
)

type ProxyConfig struct {
	ForwardURL         string
	Port               int
	HttpClientFactory  func(tr *http.Transport) *http.Client
	HttpRequestFactory func(method string, url string, body io.Reader) (*http.Request, error)
	SystemDomain       string
	ProviderId         string
	ConsumerId         string
	LoadBalancerPort   int
	IstioDirectory     string
	IpAddress          string
}

var (
	ProxyConfiguration = ProxyConfig{
		HttpClientFactory:  httpClientFactory,
		HttpRequestFactory: httpRequestFactory,
		IstioDirectory:     os.TempDir(),
		Port:               DefaultPort,
		IpAddress:          "127.0.0.1",
	}
)

type osbProxy struct {
	*http.Client
	interceptor ServiceBrokerInterceptor
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
	response, err := credentials.Update(body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(response)
}

func writeIstioFilesForProvider(istioDirectory string, bindingId string, request *model.BindRequest, response *model.BindResponse) error {
	return writeIstioConfigFiles(istioDirectory, bindingId, config.CreateIstioConfigForProvider(request, response, bindingId, ProxyConfiguration.SystemDomain))
}

func writeIstioConfigFiles(istioDirectory string, fileName string, configuration []istioModel.Config) error {
	ymlPath := path.Join(istioDirectory, fileName) + ".yml"
	log.Printf("PATH to istio config: %v\n", ymlPath)
	file, err := os.Create(ymlPath)
	if nil != err {
		return err
	}
	defer file.Close()

	fileContent, err := config.ToYamlDocuments(configuration)
	if nil != err {
		return err
	}
	_, err = file.Write([]byte(fileContent))
	if nil != err {
		return err
	}
	return nil
}

func (client osbProxy) forward(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	url := createNewUrl(ProxyConfiguration.ForwardURL, request)
	proxyRequest, err := ProxyConfiguration.HttpRequestFactory(request.Method, url, request.Body)
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
	proxyRequest := createForwardingRequest(request, err, requestBody)

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

func createForwardingRequest(request *http.Request, err error, body []byte) *http.Request {
	url := createNewUrl(ProxyConfiguration.ForwardURL, request)
	proxyRequest, err := ProxyConfiguration.HttpRequestFactory(request.Method, url, bytes.NewReader(body))
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

func SetupRouter() *gin.Engine {
	mux := gin.Default()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var interceptor ServiceBrokerInterceptor
	if ProxyConfiguration.ConsumerId != "" {
		interceptor = consumer_interceptor{}
	} else if ProxyConfiguration.ProviderId != "" {
		interceptor = producer_interceptor{}
	} else {
		interceptor = noOpInterceptor{}
	}
	client := osbProxy{ProxyConfiguration.HttpClientFactory(tr), interceptor}
	if ProxyConfiguration.ProviderId != "" {
		writeIstioConfigFiles(ProxyConfiguration.IstioDirectory, "istio-broker",
			config.CreateEntriesForExternalService("istio-broker", string(ProxyConfiguration.IpAddress), uint32(ProxyConfiguration.Port), "istio-broker."+ProxyConfiguration.SystemDomain, "client.istio.sapcloud.io", 9000))
		mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardBindRequest)
	mux.NoRoute(client.forward)

	return mux
}
