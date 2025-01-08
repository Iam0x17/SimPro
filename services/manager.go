package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"ProtoSimService/common"
)

// ServiceManager结构体，管理多个模拟服务
type ServiceManager struct {
	services []common.MockService
}

// AddService方法用于添加模拟服务到服务管理器
func (s *ServiceManager) AddService(service common.MockService) {
	s.services = append(s.services, service)
}

// StartAllServices方法启动所有添加的模拟服务
func (s *ServiceManager) StartAllServices(ctx context.Context) {
	for _, service := range s.services {
		go func(svc common.MockService) {
			addr := fmt.Sprintf(":%d", getPortForService(svc.GetServiceName()))
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				log.Printf("启动服务 %s 监听端口失败: %v\n", svc.GetServiceName(), err)
				return
			}
			log.Printf("服务 %s 启动，监听端口 %d\n", svc.GetServiceName(), getPortForService(svc.GetServiceName()))

			if svc.NeedsListener() {
				go svc.ServeWithListener(ctx, listener)
			} else {
				for {
					conn, err := listener.Accept()
					if err != nil {
						log.Printf("服务 %s 接受连接出错: %v\n", svc.GetServiceName(), err)
						continue
					}
					go svc.Serve(ctx, conn)
				}
			}
		}(service)
	}
}

// 根据服务名称获取对应的监听端口
func getPortForService(serviceName string) int {
	switch serviceName {
	case "SSH":
		return 2222
	case "FTP":
		return 2121
	case "Redis":
		return 6379
	default:
		return 0
	}
}
