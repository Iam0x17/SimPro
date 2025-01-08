package common

import (
	"context"
	"net"
)

// MockService接口，所有模拟服务都要实现该接口
type MockService interface {
	Serve(ctx context.Context, conn net.Conn)
	ServeWithListener(ctx context.Context, listener net.Listener)
	GetServiceName() string
	NeedsListener() bool
}
