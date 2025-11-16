package dto

// Общий формат ошибок, соответствующий ErrorResponse из OpenAPI.

type ErrorDTO struct {
	Error ErrorContent `json:"error"`
}

type ErrorContent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}


