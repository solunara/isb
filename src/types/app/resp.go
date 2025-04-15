package app

// Error 自定义错误类型
type ResponseType struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Response(code int, message string, data interface{}) ResponseType {
	return ResponseType{
		Code: code,
		Msg:  message,
		Data: data,
	}
}

func ResponseErr(code int, message string) ResponseType {
	return ResponseType{
		Code: code,
		Msg:  message,
	}
}

func ResponseOK(data any) ResponseType {
	return ResponseType{
		Code: 200,
		Msg:  "ok",
		Data: data,
	}
}
