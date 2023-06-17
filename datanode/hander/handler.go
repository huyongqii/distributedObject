package hander

import (
	. "com.mgface.disobj/datanode/ops"
	"net/http"
)

func ApiHandler(writer http.ResponseWriter, req *http.Request) {
	method := req.Method
	if method == http.MethodPut {
		Put(writer, req)
		return
	}
	if method == http.MethodGet {
		Get(writer, req)
		return
	}
	writer.WriteHeader(http.StatusMethodNotAllowed)
}
