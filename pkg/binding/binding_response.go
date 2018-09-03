package binding

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type networkData struct {
	profileId string
	data      nwData
}

type nwData struct {
	providerId string
	endpoints  map[string]string
}

func addIstioDataToResponse(request http.Request, body io.ReadCloser) (http.Response, error) {
	response := http.Response{Status: "200 OK"}
	fmt.Printf("response: %v", response)
	//responseBody, err := ioutil.ReadAll(response.Body)
	//if nil != err {
	//	return http.Response{}, err
	//}
	var responseBody []byte

	newBody, err := endpoints.GenerateEndpoint(responseBody)
	if nil != err {
		return http.Response{}, err
	}
	//todo call joinedNetworkData and add its return values to response

	fmt.Printf("newBody: %v", newBody)
	return response, err
}

func joinNetworkData(endpointsBody []byte) (networkData, error) {
	//"network_data": {
	//	"network_profile_id": "urn:istio:public",
	//		"data": {
	//		"provider_id": "spiffe://ingress.services.cf.dev01.aws.istio.sapcloud.io",
	//			"endpoints": [{
	//		"host": "postgres.services.cf.dev01.aws.istio.sapcloud.io",
	//		"port": 9000
	//		}]
	//	}

	endpoints := make(map[string]string)
	host, port := extractDataFromEndpoints(endpointsBody)
	endpoints["host"] = host
	endpoints["port"] = port
	providerId := "spiffe://ingress.services.cf.dev01.aws.istio.sapcloud.io"
	var data nwData
	data.endpoints = endpoints
	data.providerId = providerId

	profileId := "urn:istio:public"

	var joinedNetworkData networkData
	joinedNetworkData.profileId = profileId
	joinedNetworkData.data = data

	//var providerId, epHost, epPort string
	return joinedNetworkData, nil
}

func extractDataFromEndpoints(endpointsBody []byte) (string, string) {
	//todo adjust so that all ports and host can be found!
	var host, port string
	bodyString := string(endpointsBody[:])

	reHost := regexp.MustCompile(`addresses: \"[a-zA-Z0-9\.]*\"`)
	endpointsinformation := reHost.FindString(bodyString)
	host = strings.Split(endpointsinformation, `"`)[1]

	//todo consider `port` is a random name which can be different especially when several ports are given
	rePort := regexp.MustCompile(`port: (?P<epport>[a-zA-Z0-9\.]*)`)
	epi := rePort.FindStringSubmatch(bodyString)
	port = epi[1]

	return host, port
}
