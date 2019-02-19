package router

import (
	"crypto/tls"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

const (
	DefaultPort = 8080
)

type RouterConfig struct {
	ForwardURL         string
	SkipVerifyTLS      bool
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
	response, err := model.Adapt(request.Credentials, request.EndpointMappings)
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (client osbProxy) forward(ctx *gin.Context) {
	client.forwardWithCallback(ctx, func(ctx *gin.Context) error {
		return nil
	})
}

func (client osbProxy) deleteBinding(ctx *gin.Context) {
	bindingId := ctx.Params.ByName("binding_id")
	osbClient := InterceptedOsbClient{&OsbClient{&RouterRestClient{client.Client, ctx.Request, client.config}}, client.interceptor}
	err := osbClient.Unbind(bindingId)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{})
}

func (client osbProxy) forwardCatalog(ctx *gin.Context) {
	osbClient := InterceptedOsbClient{&OsbClient{&RouterRestClient{client.Client, ctx.Request, client.config}}, client.interceptor}
	catalog, err := osbClient.GetCatalog()
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
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
			httpError(ctx, err, http.StatusInternalServerError)
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

	osbClient := InterceptedOsbClient{&OsbClient{&RouterRestClient{client.Client, request, client.config}}, client.interceptor}
	log.Printf("Received request: %v %v", request.Method, request.URL.Path)
	bindingId := ctx.Params.ByName("binding_id")
	bindResponse, err := osbClient.Bind(bindingId, &bindRequest)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
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
	return fmt.Sprintf("%s%s", newBaseUrl, createNewPath(req))
}

func createNewPath(req *http.Request) string {
	path := req.URL.Path

	if req.URL.RawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, req.URL.RawQuery)
	}

	return path
}

func registerConsumerRelevantRoutes(prefix string, mux *gin.Engine, client *osbProxy) {
	mux.PUT(prefix+"/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardBindRequest)
	mux.DELETE(prefix+"/v2/service_instances/:instance_id/service_bindings/:binding_id", client.deleteBinding)
	mux.GET(prefix+"/v2/catalog", client.forwardCatalog)
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
		TLSClientConfig: &tls.Config{InsecureSkipVerify: routerConfig.SkipVerifyTLS},
	}
	client := osbProxy{routerConfig.HttpClientFactory(tr), interceptor, routerConfig}
	mux.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	if interceptor.HasAdaptCredentials() {
		mux.POST("/v2/service_instances/:instance_id/service_bindings/:binding_id/adapt_credentials", client.updateCredentials)
	}
	registerConsumerRelevantRoutes("", mux, &client)
	registerConsumerRelevantRoutes("/v1/osb/:broker_id", mux, &client)
	mux.NoRoute(client.forward)

	return mux
}
