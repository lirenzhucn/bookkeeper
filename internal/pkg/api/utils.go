package api

import (
	"net/http"

	"go.uber.org/zap"
)

func checkErr(err error, w http.ResponseWriter, statusCode int,
	msg string, a ...interface{}) bool {
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	if err != nil {
		a = append(a, "error")
		a = append(a, err)
		sugar.Errorw(msg, a...)
		var outMsg = msg
		if statusCode == 500 {
			outMsg = "Internal Server Error"
		}
		http.Error(w, outMsg, statusCode)
		return false
	}
	return true
}
