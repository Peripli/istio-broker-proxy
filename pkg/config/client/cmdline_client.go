package main

import (
	"bufio"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"os"
	"strconv"
	"strings"
)

func main() {
	serviceName := readStringParam("name of service")
	endpointServiceEntry := readStringParam("endpoint(ip) of service entry")
	portServiceEntryAsString := readStringParam("port of service entry")
	portServiceEntry, _ := strconv.ParseUint(portServiceEntryAsString, 10, 32)
	hostVirtualService := readStringParam("hostname of virtual service")

	out, err := config.CreateEntriesForExternalService(serviceName, endpointServiceEntry, uint32(portServiceEntry), hostVirtualService)

	if err == nil {
		fmt.Printf("%s", out)
	} else {
		fmt.Printf("error occured: %s", err.Error())
		os.Exit(1)
	}
}

func readStringParam(name string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(name + ": ")
	readParam, _ := reader.ReadString('\n')
	return strings.TrimSuffix(readParam, "\n")
}
