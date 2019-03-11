package api

type RestResponse interface {
	Into(response interface{}) error
	Error() error
}

type RestRequest interface {
	AppendPath(path string) RestRequest
	Do() RestResponse
}

type RestClient interface {
	Get() RestRequest
	Post(body interface{}) RestRequest
	Put(body interface{}) RestRequest
	Delete() RestRequest
}
