package gowebkit

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ResultOk() Result {
	return Result{Code: 0, Msg: "ok"}
}

func ResultOkData(data interface{}) Result {
	return Result{Code: 0, Msg: "ok", Data: data}
}

func ResultError(msg string) Result {
	return Result{Code: -1, Msg: msg}
}
