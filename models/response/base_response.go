package response

type BaseResponse struct {
	Meta struct {
		Status int
	}
	Data interface{}
}
