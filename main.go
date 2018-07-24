package main

import (
	"fmt"
	"net/http"
	"os"
)

const (
	DEFAULT_PORT = "8080"
)

var (
	port string
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func main() {

	if port = os.Getenv("PORT"); len(port) == 0 {
		port = DEFAULT_PORT
	}

	http.HandleFunc("/", helloWorld)
	http.ListenAndServe(":"+port, nil)
}
