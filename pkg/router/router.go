package router

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
		httpError(ctx, err, http.StatusBadRequest)
		return
	}
	if len(request.EndpointMappings) == 0 {
		httpError(ctx, errors.New("No endpoint mappings available"), http.StatusBadRequest)
		return
	}
	response, err := model.Adapt(request.Credentials, request.EndpointMappings)
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

type OsbRequest struct {
	url     string
	err     error
	request []byte
	method  string
	header  http.Header
	client  osbProxy
}

type OsbResponse struct {
	err      error
	response []byte
	url      string
}

func (client osbProxy) Get() *OsbRequest {
	return &OsbRequest{method: http.MethodGet, client: client, request: make([]byte, 0)}
}

func (client osbProxy) Post(request interface{}) *OsbRequest {
	requestBody, err := json.Marshal(request)
	return &OsbRequest{method: http.MethodPost, client: client, request: requestBody, err: err}
}

func (client osbProxy) Put(request interface{}) *OsbRequest {
	requestBody, err := json.Marshal(request)
	return &OsbRequest{method: http.MethodPut, client: client, request: requestBody, err: err}
}

func (o *OsbRequest) Header(header http.Header) *OsbRequest {
	o.header = header
	return o
}

func (o *OsbRequest) Url(url string) *OsbRequest {
	o.url = url
	return o
}

func (o *OsbRequest) Do() *OsbResponse {
	osbResponse := OsbResponse{err: o.err, url: o.url}
	if o.err != nil {
		return &osbResponse
	}
	var proxyRequest *http.Request
	proxyRequest, osbResponse.err = o.client.config.HttpRequestFactory(o.method, o.url, bytes.NewReader(o.request))
	if osbResponse.err != nil {
		log.Printf("ERROR: %s\n", osbResponse.err.Error())
		return &osbResponse
	}
	proxyRequest.Header = o.header

	var response *http.Response
	response, osbResponse.err = o.client.Do(proxyRequest)
	if osbResponse.err != nil {
		log.Printf("ERROR: %s\n", osbResponse.err.Error())
		return &osbResponse
	}
	log.Printf("Response status from %s: %s\n", o.url, response.Status)

	defer response.Body.Close()

	osbResponse.response, osbResponse.err = ioutil.ReadAll(response.Body)
	if nil != osbResponse.err {
		log.Printf("ERROR: %s\n", osbResponse.err.Error())
		return &osbResponse
	}

	osbResponse.err = model.HttpErrorFromResponse(response.StatusCode, osbResponse.response)
	if osbResponse.err != nil {
		return &osbResponse
	}

	return &osbResponse
}

func (o *OsbResponse) Into(result interface{}) error {
	if o.err != nil {
		return o.err
	}
	o.err = json.Unmarshal(o.response, result)

	if nil != o.err {
		o.err = fmt.Errorf("Can't unmarshal response from %s: %s", o.url, o.err.Error())
		log.Printf("ERROR: %s\n", o.err.Error())
		return o.err
	}
	return nil
}

func (o *OsbResponse) Error() error {
	return o.err
}

func (client osbProxy) getCatalog(header http.Header) (*model.Catalog, error) {
	var catalog model.Catalog
	err := client.Get().
		Url(fmt.Sprintf("%s/v2/catalog", client.config.ForwardURL)).
		Header(header).
		Do().
		Into(&catalog)
	return &catalog, err

}

func (client osbProxy) adaptCredentials(credentials model.Credentials, mapping []model.EndpointMapping, instanceId string, bindId string, header http.Header) (*model.BindResponse, error) {

	var bindResponse model.BindResponse
	err := client.Post(&model.AdaptCredentialsRequest{Credentials: credentials, EndpointMappings: mapping}).
		Url(fmt.Sprintf("%s/v2/service_instances/%s/service_bindings/%s/adapt_credentials", client.config.ForwardURL, instanceId, bindId)).
		Header(header).
		Do().
		Into(&bindResponse)
	return &bindResponse, err

}

func (client osbProxy) forward(ctx *gin.Context) {
	client.forwardWithCallback(ctx, func(ctx *gin.Context) error {
		return nil
	})
}

func (client osbProxy) deleteBinding(ctx *gin.Context) {
	client.forwardWithCallback(ctx, func(ctx *gin.Context) error {
		return client.interceptor.PostDelete(ctx.Params.ByName("binding_id"))
	})
}

func (client osbProxy) forwardCatalog(ctx *gin.Context) {
	catalog, err := client.getCatalog(ctx.Request.Header)
	if err != nil {
		httpError(ctx, err, http.StatusBadGateway)
		return
	}
	client.interceptor.PostCatalog(catalog)
	ctx.JSON(http.StatusOK, catalog)
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
		httpError(ctx, err, http.StatusBadGateway)
		return
	}
	log.Printf("Request forwarded %v: %s\n", request.URL, response.Status)

	defer response.Body.Close()

	if (response.StatusCode / 100) == 2 {
		err = postCallback(ctx)
		if err != nil {
			httpError(ctx, err, http.StatusBadGateway)
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
	request := ctx.Request

	var bindRequest model.BindRequest
	err := ctx.ShouldBindJSON(&bindRequest)
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}

	log.Printf("Received request: %v %v", request.Method, request.URL.Path)

	bindRequest = *client.interceptor.PreBind(bindRequest)

	var bindResponse model.BindResponse

	err = client.Put(bindRequest).
		Header(request.Header).
		Url(createNewUrl(client.config.ForwardURL, request)).
		Do().
		Into(&bindResponse)

	if err != nil {
		httpError(ctx, err, http.StatusBadGateway)
		return
	}

	bindingId := ctx.Params.ByName("binding_id")
	instanceId := ctx.Params.ByName("instance_id")
	modifiedBindResponse, err := client.interceptor.PostBind(bindRequest, bindResponse, bindingId,
		func(credentials model.Credentials, mappings []model.EndpointMapping) (*model.BindResponse, error) {
			return client.adaptCredentials(credentials, mappings, instanceId, bindingId, request.Header)
		})
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
	bindResponse = *modifiedBindResponse

	ctx.JSON(http.StatusOK, bindResponse)

}

func httpError(ctx *gin.Context, err error, statusCode int) {
	log.Printf("ERROR: %s\n", err.Error())
	httpError := model.HttpErrorFromError(err, statusCode)
	ctx.AbortWithStatusJSON(httpError.StatusCode, httpError)
}

func httpClientFactory(tr *http.Transport) *http.Client {
	client := &http.Client{Transport: tr}
	return client
}

func httpRequestFactory(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
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
	if interceptor.HasAdaptCredentials() {
		mux.POST("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	mux.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardBindRequest)
	mux.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", client.deleteBinding)
	mux.GET("/v2/catalog", client.forwardCatalog)
	mux.NoRoute(client.forward)

	return mux
}
