package telnet

import (
	"ProtoSimService/common"
	"context"
	"github.com/globalcyberalliance/telnet-go"
	"github.com/globalcyberalliance/telnet-go/shell"
	"github.com/sirupsen/logrus"
	"net"
)

var telnetLogger *logrus.Logger

func init() {
	telnetLogger = common.SetupServiceLogger("Telnet", true)
}

// MockTelnetService 实现通用的MockService接口
type SimTelnetService struct{}

func (m *SimTelnetService) NeedsListener() bool {
	return true
}

func (m *SimTelnetService) Serve(ctx context.Context, conn net.Conn) {

}

// Serve方法处理FTP连接相关逻辑
func (m *SimTelnetService) ServeWithListener(ctx context.Context, listener net.Listener) {
	authHandler := shell.NewAuthHandler("root", "123456", 3)
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
	err := telnet.Serve(listener, srv.HandlerFunc)
	if err != nil {
		return
	}
}

// GetServiceName方法返回服务名称
func (m *SimTelnetService) GetServiceName() string {
	return "Telnet"
}
