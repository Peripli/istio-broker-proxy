package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/credentials"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
)

const (
	DefaultPort      = "8080"
	ServiceFabrikURL = "10.11.252.10:9293/cf"
)

type ProxyConfig struct {
	ForwardURL string
	port       string
}

var (
	config = ProxyConfig{ServiceFabrikURL, DefaultPort}
	log    = make([]string, 0)
)

func update_credentials(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	log = append(log, "Received update: "+string(body))
	response, err := credentials.Update(body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(response)
}

func info(w http.ResponseWriter, r *http.Request) {
	for _, line := range log {
		fmt.Fprintf(w, "%s\n\n", line)
	}
}

func translateBody(originalRequest *http.Request, responseBody []byte) ([]byte, error) {
	newBody, err := endpoints.GenerateEndpoint(responseBody)
	return newBody, err
}

func createNewUrl(newHost string, req *http.Request) string {

	url := fmt.Sprintf("https://%s%s", newHost, req.URL.Path)

	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.URL.RawQuery)
	}

	return url
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log = append(log, "Received redirect: "+string(body))

	// create a new url from the raw RequestURI sent by the client
	url := createNewUrl(config.ForwardURL, req)
	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// reassign the body for the dump
	proxyReq.Body = ioutil.NopCloser(bytes.NewReader(body))
	requestDump, err := httputil.DumpRequest(proxyReq, true)
	if err != nil {
		log = append(log, err.Error())
	}
	log = append(log, "Request: "+string(requestDump))

	defer func() {
		resp.Body.Close()
	}()

	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err = translateBody(req, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)

	// reassign body for dump
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log = append(log, err.Error())
	}

	log = append(log, "Response: "+string(responseDump))

}

func readPort() {
	port := os.Getenv("PORT")
	if len(port) != 0 {
		config.port = port
	} else {
		config.port = DefaultPort
	}
}

func main() {
	readPort()
	http.HandleFunc("/info", info)
	http.HandleFunc("/update_credentials", update_credentials)
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":"+config.port, nil)
}
