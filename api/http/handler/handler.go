package handler

import (
	"SimPro/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceControlRequest struct {
	ServiceName string `json:"service_name"`
	Action      string `json:"action"`
}

type ServiceConfig struct {
	ServiceName string                 `json:"service_name"`
	Config      map[string]interface{} `json:"config"`
}

func ServiceControlsHandler(c *gin.Context) {
	var err error
	serviceName := c.Query("service_name")
	action := c.Query("action")

	if serviceName == "" || action == "" {
		// 如果URL参数为空，尝试从JSON请求体中获取
		var req ServiceControlRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing service_name or action parameter"})
			return
		}
		serviceName = req.ServiceName
		action = req.Action
	}

	manager := services.GetServiceManager()

	switch action {
	case "start":
		err = manager.StartServiceByName(serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start service %s: %v", serviceName, err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"message": fmt.Sprintf("Service %s started", serviceName),
		})
	case "stop":
		err = manager.StopServiceByName(serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to stop service %s: %v", serviceName, err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"message": fmt.Sprintf("Service %s stopped", serviceName),
		})
	case "status":
		var status string
		status, err = manager.GetServiceStatusByName(serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get status for service %s: %v", serviceName, err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"status":  status,
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Use 'start' or 'stop'"})
	}
}

func ServiceConfigsHandler(c *gin.Context) {
	manager := services.GetServiceManager()

	if c.Request.Method == "GET" {

		// 获取所有服务配置
		allConfigs := make(map[string]map[string]interface{})
		services := []string{"ssh", "redis", "mysql", "postgres", "telnet", "ftp"}
		for _, svc := range services {
			config, err := manager.GetServiceConfig(svc)
			if err != nil {
				continue
			}
			allConfigs[svc] = config
		}
		c.JSON(http.StatusOK, allConfigs)
		return

	} else if c.Request.Method == "POST" {
		// 更新服务配置
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		serviceName := req["service_name"].(string)
		// 确保serviceName不为空
		if serviceName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing service_name parameter"})
			return
		}

		err := manager.UpdateServiceConfig(serviceName, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update config for service %s: %v", serviceName, err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"message": "Config updated successfully",
		})
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	}
}
