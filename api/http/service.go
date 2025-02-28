package http

import (
	"SimPro/api/http/handler"
	"SimPro/common"
	"fmt"
	"net/http"
)

func StartHttpService() error {
	http.HandleFunc("/service-control", handler.ServiceControlHandler)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			common.Logger.Error(fmt.Sprintf("http service start error: %v", err))
		} else {
			common.Logger.Info("http service start success")
		}

	}()
	return nil
}
