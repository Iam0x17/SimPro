package http

import (
	"SimPro/api/http/handler"
	"fmt"
	"net/http"
)

func StartHttpService() error {
	http.HandleFunc("/service-control", handler.ServiceControlHandler)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("http service start error", err)
		}
		fmt.Println("http service start success")
	}()
	return nil
}
