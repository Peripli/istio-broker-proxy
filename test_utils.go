package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func setupRouterStub() *gin.Engine {

	mux := gin.Default()
	mux.NoRoute(mirrorRequest)

	return mux
}

func mirrorRequest(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}
