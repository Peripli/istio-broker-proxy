package router

import (
	"crypto/tls"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/gin-gonic/gin"
	"io"
	"istio.io/istio/pkg/log"
	"net/http"
)

const (
	// DefaultPort for istio-broker-proxy HTTP endpoint
	DefaultPort        = 8080
	// IstioBrokerVersion is a header entry that contains the commit-shas of istio-broker-proxy
	IstioBrokerVersion = "X-Istio-Broker-Versions"

	healthEnpoint = "/health"
)

var version string

//Config contains various config
type Config struct {
	ForwardURL         string
	SkipVerifyTLS      bool
	Port               int
	HTTPClientFactory  func(tr *http.Transport) *http.Client
	HTTPRequestFactory func(method string, url string, header http.Header, body io.Reader) (*http.Request, error)
}

type osbProxy struct {
	*http.Client
	interceptor ServiceBrokerInterceptor
	config      Config
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

func (client osbProxy) forwardUnbindRequest(ctx *gin.Context) {
	bindingID := ctx.Params.ByName("binding_id")
	osbClient := interceptedOsbClient{&osbClient{&restClient{client.Client, ctx.Request, client.config}}, client.interceptor}
	err := osbClient.Unbind(bindingID)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{})
}

func (client osbProxy) forwardCatalog(ctx *gin.Context) {
	osbClient := interceptedOsbClient{&osbClient{&restClient{client.Client, ctx.Request, client.config}}, client.interceptor}
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

	log.Infof("Received request: %v %v", request.Method, request.URL.Path)

	url := createNewURL(client.config.ForwardURL, request)
	proxyRequest, err := client.config.HTTPRequestFactory(request.Method, url, request.Header, request.Body)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
	response, err := client.Do(proxyRequest)
	if err != nil {
		httpError(ctx, err, http.StatusBadGateway)
		return
	}
	log.Infof("Request forwarded %v: %s\n", request.URL, response.Status)

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

	osbClient := interceptedOsbClient{&osbClient{&restClient{client.Client, request, client.config}}, client.interceptor}
	log.Infof("Received request: %v %v", request.Method, request.URL.Path)
	bindingID := ctx.Params.ByName("binding_id")
	bindResponse, err := osbClient.Bind(bindingID, &bindRequest)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, bindResponse)
}

func (client osbProxy) forwardProvisionRequest(ctx *gin.Context) {
	request := ctx.Request

	var provisionRequest model.ProvisionRequest
	err := ctx.ShouldBindJSON(&provisionRequest)
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}

	osbClient := interceptedOsbClient{&osbClient{&restClient{client.Client, request, client.config}}, client.interceptor}
	log.Infof("Received request: %v %v", request.Method, request.URL.Path)
	provisionResponse, err := osbClient.Provision(&provisionRequest)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, provisionResponse)
}

func httpError(ctx *gin.Context, err error, statusCode int) {
	log.Errorf("ERROR: %s\n", err.Error())
	httpError := model.HTTPErrorFromError(err, statusCode)
	ctx.AbortWithStatusJSON(httpError.StatusCode, httpError)
}

func httpClientFactory(tr *http.Transport) *http.Client {
	client := &http.Client{Transport: tr}
	return client
}

func httpRequestFactory(method string, url string, header http.Header, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Errorf("error during create request: %s\n", err.Error())
		return nil, err
	}
	if header != nil {
		request.Header = header
	} else {
		request.Header = http.Header{}
	}
	request.Header.Add(IstioBrokerVersion, version)
	log.Infof("Added header %s with value %s\n", IstioBrokerVersion, version)

	return request, nil
}

func createNewURL(newBaseURL string, req *http.Request) string {
	return fmt.Sprintf("%s%s", newBaseURL, createNewPath(req))
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
	mux.DELETE(prefix+"/v2/service_instances/:instance_id/service_bindings/:binding_id", client.forwardUnbindRequest)
	mux.PUT(prefix+"/v2/service_instances/:instance_id", client.forwardProvisionRequest)
	mux.GET(prefix+"/v2/catalog", client.forwardCatalog)
}

func logAndAddVersionHeader(ctx *gin.Context) {
	path := ctx.Request.URL.Path
	istioBrokerVersion := ctx.Request.Header.Get(IstioBrokerVersion)
	ctx.Next()
	ctx.Header(IstioBrokerVersion, version)
	if path != healthEnpoint {
		log.Infof("Header %s:  received \"%s\", responded \"%s\"\n",IstioBrokerVersion, istioBrokerVersion, version)
	}
}

//SetupRouter creates the istio-broker-proxy's endpoints
func SetupRouter(interceptor ServiceBrokerInterceptor, routerConfig Config) *gin.Engine {
	return SetupRouterWithVersion(interceptor,routerConfig,"")
}
//SetupRouterWithVersion creates the istio-broker-proxy's endpoints
func SetupRouterWithVersion(interceptor ServiceBrokerInterceptor, routerConfig Config, vers string) *gin.Engine {
	version = vers
	if routerConfig.HTTPClientFactory == nil {
		routerConfig.HTTPClientFactory = httpClientFactory
	}
	if routerConfig.HTTPRequestFactory == nil {
		routerConfig.HTTPRequestFactory = httpRequestFactory
	}
	if routerConfig.Port == 0 {
		routerConfig.Port = DefaultPort
	}

	mux := gin.New()
	mux.Use(gin.LoggerWithWriter(gin.DefaultWriter, healthEnpoint), gin.Recovery())
	mux.Use(logAndAddVersionHeader)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: routerConfig.SkipVerifyTLS},
	}
	client := osbProxy{routerConfig.HTTPClientFactory(tr), interceptor, routerConfig}
	mux.GET(healthEnpoint, func(ctx *gin.Context) {
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
