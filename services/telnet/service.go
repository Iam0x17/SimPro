package telnet

import (
	"SimPro/common"
	"SimPro/config"
	"fmt"
	"github.com/globalcyberalliance/telnet-go"
	"github.com/globalcyberalliance/telnet-go/shell"
	"go.uber.org/zap"
	"net"
	"sync"
)

// MockTelnetService 实现通用的MockService接口
type SimTelnetService struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func (s *SimTelnetService) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
	}
	s.wg.Wait()
	common.Logger.Info(common.EventStopService, zap.String("protocol", "telnet"), zap.String("info", "Telnet service has stopped"))
	return nil
}

// Serve方法处理FTP连接相关逻辑
func (s *SimTelnetService) Start(cfg *config.Config) error {

	var err error
	s.listener, err = net.Listen("tcp", ":"+cfg.Telnet.Port)
	if err != nil {
		return err
	}
	//telnetLogger.Printf("Telnet 服务正在监听端口 %s", cfg.Telnet.Port)
	common.Logger.Info(common.EventStartService, zap.String("protocol", "telnet"), zap.String("info", fmt.Sprintf("Telnet service is listening on port %s", cfg.Telnet.Port)))
	authHandler := shell.NewAuthHandler(cfg.Telnet.User, cfg.Telnet.Pass, 3)
	commands := []shell.Command{
		{
			Regex:    "^docker$",
			Response: "\nUsage:  docker [OPTIONS] COMMAND\r\n",
		},
		{
			Regex:    "^docker .*$",
			Response: "Error response from daemon: dial unix docker.raw.sock: connect: connection refused\r\n",
		},
		{
			Regex:    "^uname$",
			Response: "Linux\r\n",
		},
	}

	srv := shell.Server{AuthHandler: authHandler, Commands: commands}
	go telnet.Serve(s.listener, srv.HandlerFunc)

	return nil
}

// GetServiceName方法返回服务名称
func (m *SimTelnetService) GetName() string {
	return "telnet"
}
