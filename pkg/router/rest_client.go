package router

type RestResponse interface {
	Into(response interface{}) error
	Error() error
}

type RestRequest interface {
	Path(path string) RestRequest
	Do() RestResponse
}

type RestClient interface {
	Get() RestRequest
	Post(body interface{}) RestRequest
	Put(body interface{}) RestRequest
	Delete() RestRequest
}
