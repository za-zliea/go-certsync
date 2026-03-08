package client

const (
	SUCCESS = iota
	ERROR
)

type ResponseDTO struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	status  int         `json:"-"`
}

func (r *ResponseDTO) IsSuccess() bool {
	return r.Code == SUCCESS
}

func Success() ResponseDTO {
	return ResponseDTO{Code: SUCCESS, status: 200, Message: "SUCCESS"}
}

func SuccessWithD(data interface{}) ResponseDTO {
	return ResponseDTO{Code: SUCCESS, status: 200, Message: "SUCCESS", Data: data}
}

func SuccessWithM(message string) ResponseDTO {
	return ResponseDTO{Code: SUCCESS, status: 200, Message: message}
}

func SuccessWithDM(data interface{}, message string) ResponseDTO {
	return ResponseDTO{Code: SUCCESS, status: 200, Message: message, Data: data}
}

func Failed(message string) ResponseDTO {
	return ResponseDTO{Code: ERROR, status: 500, Message: message}
}

func FailedWithS(message string, status int) ResponseDTO {
	return ResponseDTO{Code: ERROR, status: status, Message: message}
}

func FailedWithD(message string, data interface{}) ResponseDTO {
	return ResponseDTO{Code: ERROR, status: 500, Message: message, Data: data}
}

func FailedWithDS(message string, data interface{}, status int) ResponseDTO {
	return ResponseDTO{Code: ERROR, status: status, Message: message, Data: data}
}
