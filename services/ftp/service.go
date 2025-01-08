package ftp

import (
	"context"
	"fmt"
	"log"
	"net"

	"ProtoSimService/config"
	"ProtoSimService/services/ftp/auth"
	"ProtoSimService/services/ftp/driver"

	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
)

// MockFTPService结构体实现MockService接口
type MockFTPService struct{}

func (m *MockFTPService) NeedsListener() bool {
	return true
}

func (m *MockFTPService) Serve(ctx context.Context, conn net.Conn) {

}

// Serve方法处理FTP连接相关逻辑
func (m *MockFTPService) ServeWithListener(ctx context.Context, listener net.Listener) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 定义FTP服务器驱动
	var ftpDriver server.Driver
	ftpDriver, err = file.NewDriver("./")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 定义FTP服务器认证方式
	userPass := &auth.UserPass{
		User:  cfg.FTP.User,
		Pass:  cfg.FTP.Pass,
		IsSet: true,
	}
	var serverAuth server.Auth
	if userPass.GetIsSet() {
		serverAuth = &server.SimpleAuth{Name: userPass.GetUser(), Password: userPass.GetPassword()}
	} else {
		serverAuth = &auth.ZeroAuth{}
	}

	// 组装FTP服务器选项
	opt := &server.Options{
		Name:           "iwebd",
		Driver:         ftpDriver,
		Port:           cfg.FTP.Port,
		Auth:           serverAuth,
		Perm:           driver.NewFsPerm("./"),
		WelcomeMessage: cfg.FTP.WelcomeMessage,
	}

	s, err := server.NewServer(opt)
	if err != nil {
		log.Fatalf("Failed to serve ftp: %v", err)
		return
	}

	// 运行服务器
	quitNotifier := make(chan int)
	go func() {
		err = s.Serve(listener)
		if err != nil && err != server.ErrServerClosed {
			log.Printf("Failed to serve ftp: %v", err)
		}
		quitNotifier <- 0
	}()

	// 等待服务器结束
	<-quitNotifier
	s.Shutdown()
}

// GetServiceName方法返回服务名称
func (m *MockFTPService) GetServiceName() string {
	return "FTP"
}
