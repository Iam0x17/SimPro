package http

import (
	"SimPro/api/http/handler"
	"SimPro/common"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func StartHttpService(port string, assetsFs embed.FS) error {
	//r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(CORSMiddleware())

	// 静态文件服务
	dist, err := fs.Sub(assetsFs, "assets/dist")
	if err != nil {
		return fmt.Errorf("无法加载静态文件: %v", err)
	}
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "admin/")
	})
	r.StaticFS("/admin", http.FS(dist))

	// API路由
	r.Any("/api/service", handler.ServiceControlsHandler)
	r.Any("/api/service/config", handler.ServiceConfigsHandler)

	// 处理前端路由
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(dist))
	})

	go func() {
		common.Logger.Info(fmt.Sprintf("Web服务正在启动，监听端口 :%s", port))
		err := r.Run(":" + port)
		if err != nil {
			common.Logger.Error(fmt.Sprintf("http service start error: %v", err))
		} else {
			common.Logger.Info("http service start success")
		}
	}()
	return nil
}
