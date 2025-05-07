package app

// Error 自定义错误类型
type ResponseType struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type ResponsePagaDataType struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
}

type BaseMap struct {
	HosName    string `json:"hosName"`
	FatherName string `json:"fatherName"`
	Name       string `json:"name"`
}

type ResponseRegistrationPagaDataType struct {
	List    any     `json:"list"`
	BaseMap BaseMap `json:"baseMap"`
	Total   int64   `json:"total"`
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

func ResponsePageData(total int64, data any) ResponseType {
	return ResponseType{
		Code: 200,
		Msg:  "ok",
		Data: ResponsePagaDataType{
			List:  data,
			Total: total,
		},
	}
}

func ResponseRegistrationPageData(total int64, list any, baseMap BaseMap) ResponseType {
	return ResponseType{
		Code: 200,
		Msg:  "ok",
		Data: ResponseRegistrationPagaDataType{
			List:    list,
			BaseMap: baseMap,
			Total:   total,
		},
	}
}
