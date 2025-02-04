package ftp

import (
	"SimPro/common"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net"
	"strconv"
	"sync"

	"SimPro/config"
	"SimPro/services/ftp/auth"
	"SimPro/services/ftp/driver"

	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
)

type SimFTPService struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func (s *SimFTPService) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
	}
	common.Logger.Info(common.EventStopService, zap.String("protocol", "ftp"), zap.String("info", "FTP service has stopped"))
	//ftpLogger.Println("FTP 服务已停止")
	return nil
}

// Serve方法处理FTP连接相关逻辑
func (s *SimFTPService) Start(cfg *config.Config) error {

	var err error
	s.listener, err = net.Listen("tcp", ":"+cfg.FTP.Port)
	if err != nil {
		return err
	}
	common.Logger.Info(common.EventStartService, zap.String("protocol", "ftp"), zap.String("info", fmt.Sprintf("FTP service is listening on port %s", cfg.FTP.Port)))
	//ftpLogger.Printf("FTP 服务正在监听端口 %s", cfg.FTP.Port)
	// 定义FTP服务器驱动
	var ftpDriver server.Driver
	ftpDriver, err = file.NewDriver("./")
	if err != nil {
		fmt.Println(err)
		return err
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

	port, err := strconv.Atoi(cfg.Postgres.Port)
	// 组装FTP服务器选项
	opt := &server.Options{
		Name:           "iwebd",
		Driver:         ftpDriver,
		Port:           port,
		Auth:           serverAuth,
		Perm:           driver.NewFsPerm("./"),
		WelcomeMessage: "welcome ftp",
	}

	sServer, err := server.NewServer(opt)
	if err != nil {
		log.Fatalf("Failed to serve ftp: %v", err)
		return err
	}

	// 运行服务器
	quitNotifier := make(chan int)
	go func() {
		err = sServer.Serve(s.listener)
		if err != nil && err != server.ErrServerClosed {
			log.Printf("Failed to serve ftp: %v", err)
		}
		quitNotifier <- 0
	}()

	//// 等待服务器结束
	//<-quitNotifier
	//sServer.Shutdown()
	return nil
}

// GetServiceName方法返回服务名称
func (m *SimFTPService) GetName() string {
	return "ftp"
}
