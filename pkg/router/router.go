package router

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/credentials"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
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
	forwardURL         string
	port               int
	httpClientFactory  func(tr *http.Transport) *http.Client
	httpRequestFactory func(method string, url string, body io.Reader) (*http.Request, error)
	systemDomain       string
	providerId         string
	consumerId         string
	loadBalancerPort   int
	istioDirectory     string
	ipAddress          string
}

var (
	proxyConfig = ProxyConfig{
		httpClientFactory:  httpClientFactory,
		httpRequestFactory: httpRequestFactory,
		istioDirectory:     os.TempDir(),
		port:               DefaultPort,
		ipAddress:          "127.0.0.1",
	}
)

type osbProxy struct {
	*http.Client
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
	return writeIstioConfigFiles(istioDirectory, bindingId, config.CreateIstioConfigForProvider(request, response, bindingId))
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
	file.Write([]byte(fileContent))
	return nil
}

func (client osbProxy) forward(ctx *gin.Context) {
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

	if proxyConfig.consumerId != "" {
		bindRequest.NetworkData.Data.ConsumerId = proxyConfig.consumerId
		bindRequest.NetworkData.NetworkProfileId = profiles.NetworkProfile
	}

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
		if proxyConfig.providerId != "" {
			bindingId := ctx.Params.ByName("binding_id")
			systemDomain := proxyConfig.systemDomain
			providerId := proxyConfig.providerId
			if len(bindResponse.Endpoints) == 0 {
				bindResponse.Endpoints = bindResponse.Credentials.Endpoints
			}
			profiles.AddIstioNetworkDataToResponse(providerId, bindingId, systemDomain, proxyConfig.loadBalancerPort, &bindResponse)

			writeIstioFilesForProvider(proxyConfig.istioDirectory, bindingId, &bindRequest, &bindResponse)
		}

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

func SetupRouter() *gin.Engine {
	mux := gin.Default()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := osbProxy{proxyConfig.httpClientFactory(tr)}
	if proxyConfig.providerId != "" {
		writeIstioConfigFiles(proxyConfig.istioDirectory, "istio-broker",
			config.CreateEntriesForExternalService("istio-broker", string(proxyConfig.ipAddress), uint32(proxyConfig.port), "istio-broker-host", "client.istio.sapcloud.io", 9000))
		mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardBindRequest)
	mux.NoRoute(client.forward)

	return mux
}

func Run() {
	SetupConfiguration()
	flag.Parse()

	log.Printf("Running on port %d\n", proxyConfig.port)

	router := SetupRouter()
	router.Run(fmt.Sprintf(":%d", proxyConfig.port))

}

func SetupConfiguration() {
	flag.StringVar(&proxyConfig.forwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.StringVar(&proxyConfig.systemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&proxyConfig.providerId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")
	flag.StringVar(&proxyConfig.consumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")
	flag.IntVar(&proxyConfig.loadBalancerPort, "loadBalancerPort", 0, "port of the load balancer of the landscape")
	flag.StringVar(&proxyConfig.istioDirectory, "istioDirectory", os.TempDir(), "Directory to store the istio configuration files")
	flag.StringVar(&proxyConfig.ipAddress, "ipAddress", "127.0.0.1", "IP address of ingress")
	flag.IntVar(&proxyConfig.port, "port", DefaultPort, "Server listen port")
}
