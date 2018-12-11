package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"io/ioutil"
	"log"
	"net/http"
)

type RouterRestClient struct {
	*http.Client
	request *http.Request
	config  RouterConfig
}

type RouterRestRequest struct {
	url     string
	err     error
	request []byte
	method  string
	client  *RouterRestClient
}

type RouterRestResponse struct {
	err      error
	response []byte
	url      string
}

func (client *RouterRestClient) Get() RestRequest {
	return client.createRequest(http.MethodGet, make([]byte, 0), nil)
}

func (client *RouterRestClient) Delete() RestRequest {
	return client.createRequest(http.MethodDelete, make([]byte, 0), nil)
}

func (client *RouterRestClient) Post(request interface{}) RestRequest {
	requestBody, err := json.Marshal(request)
	return client.createRequest(http.MethodPost, requestBody, err)
}

func (client *RouterRestClient) Put(request interface{}) RestRequest {
	requestBody, err := json.Marshal(request)
	return client.createRequest(http.MethodPut, requestBody, err)
}

func (client *RouterRestClient) createRequest(method string, body []byte, err error) *RouterRestRequest {
	return &RouterRestRequest{method: method, client: client, request: body, url: createNewUrl(client.config.ForwardURL, client.request)}
}

func (o *RouterRestRequest) AppendPath(path string) RestRequest {
	// CAUTION: discards query
	o.url = o.client.config.ForwardURL + "/" + o.client.request.URL.Path + path
	return o
}

func (o *RouterRestRequest) Do() RestResponse {
	osbResponse := RouterRestResponse{err: o.err, url: o.url}
	if o.err != nil {
		return &osbResponse
	}
	var proxyRequest *http.Request
	proxyRequest, osbResponse.err = o.client.config.HttpRequestFactory(o.method, o.url, bytes.NewReader(o.request))
	if osbResponse.err != nil {
		log.Printf("ERROR: %s\n", osbResponse.err.Error())
		return &osbResponse
	}
	proxyRequest.Header = o.client.request.Header

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

func (o *RouterRestResponse) Into(result interface{}) error {
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

func (o *RouterRestResponse) Error() error {
	return o.err
}
