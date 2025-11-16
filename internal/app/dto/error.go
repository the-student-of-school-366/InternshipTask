package dto

type ErrorDTO struct {
	Error ErrorContent `json:"error"`
}

type ErrorContent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
