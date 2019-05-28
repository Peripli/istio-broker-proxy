package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/api"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"io/ioutil"
	"istio.io/istio/pkg/log"
	"net/http"
)

type restClient struct {
	*http.Client
	request *http.Request
	config  Config
}

type restRequest struct {
	url     string
	err     error
	request []byte
	method  string
	client  *restClient
}

type restResponse struct {
	err      error
	response []byte
	url      string
}

func (client *restClient) Get() api.RESTRequest {
	return client.createRequest(http.MethodGet, make([]byte, 0), nil)
}

func (client *restClient) Delete() api.RESTRequest {
	return client.createRequest(http.MethodDelete, make([]byte, 0), nil)
}

func (client *restClient) Post(request interface{}) api.RESTRequest {
	requestBody, err := json.Marshal(request)
	return client.createRequest(http.MethodPost, requestBody, err)
}

func (client *restClient) Put(request interface{}) api.RESTRequest {
	requestBody, err := json.Marshal(request)
	return client.createRequest(http.MethodPut, requestBody, err)
}

func (client *restClient) createRequest(method string, body []byte, err error) *restRequest {
	return &restRequest{method: method, client: client, request: body, url: createNewURL(client.config.ForwardURL, client.request)}
}

func (o *restRequest) AppendPath(path string) api.RESTRequest {
	// CAUTION: discards query
	o.url = o.client.config.ForwardURL + o.client.request.URL.Path + path
	return o
}

func (o *restRequest) Do() api.RESTResponse {
	osbResponse := restResponse{err: o.err, url: o.url}
	if o.err != nil {
		return &osbResponse
	}
	var proxyRequest *http.Request
	proxyRequest, osbResponse.err = o.client.config.HTTPRequestFactory(o.method, o.url, o.client.request.Header, bytes.NewReader(o.request))
	if osbResponse.err != nil {
		log.Errorf("error during create request: %s\n", osbResponse.err.Error())
		return &osbResponse
	}

	var response *http.Response
	response, osbResponse.err = o.client.Do(proxyRequest)
	if osbResponse.err != nil {
		log.Errorf("error during execute request: %s\n", osbResponse.err.Error())
		return &osbResponse
	}
	log.Infof("response status from %s: %s\n", o.url, response.Status)

	defer response.Body.Close()

	osbResponse.response, osbResponse.err = ioutil.ReadAll(response.Body)
	if nil != osbResponse.err {
		log.Errorf("error during read response: %s\n", osbResponse.err.Error())
		return &osbResponse
	}

	osbResponse.err = model.HTTPErrorFromResponse(response.StatusCode, osbResponse.response, o.url, o.method)
	if osbResponse.err != nil {
		return &osbResponse
	}

	return &osbResponse
}

func (o *restResponse) Into(result interface{}) error {
	if o.err != nil {
		return o.err
	}
	o.err = json.Unmarshal(o.response, result)

	if nil != o.err {
		o.err = fmt.Errorf("Can't unmarshal response from %s: %s", o.url, o.err.Error())
		log.Errorf("ERROR: %s\n", o.err.Error())
		return o.err
	}
	return nil
}

func (o *restResponse) Error() error {
	return o.err
}
