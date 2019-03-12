package api

//RESTResponse encapsulates operations for a RESTResponse
type RESTResponse interface {
	Into(response interface{}) error
	Error() error
}

//RESTRequest encapsulates operations for a RESTRequest
type RESTRequest interface {
	AppendPath(path string) RESTRequest
	Do() RESTResponse
}

//RESTClient encapsulates operations for a RESTClient
type RESTClient interface {
	Get() RESTRequest
	Post(body interface{}) RESTRequest
	Put(body interface{}) RESTRequest
	Delete() RESTRequest
}
