package redis

import (
	"SimPro/common"
	"SimPro/config"
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	redisLogger *logrus.Logger
)

func init() {
	redisLogger = common.SetupServiceLogger("Redis", true)
}

// MockRedisService结构体实现通用的MockService接口
type SimRedisService struct{}

func (m *SimRedisService) NeedsListener() bool {
	return false
}

func (m *SimRedisService) ServeWithListener(ctx context.Context, listener net.Listener) {

}

// Serve方法处理Redis连接相关逻辑
func (m *SimRedisService) Serve(ctx context.Context, conn net.Conn) {
	cfg, err := config.LoadConfig()
	if err != nil {
		redisLogger.Fatalf("加载配置失败: %v", err)
	}

	var authenticated bool
	var wg sync.WaitGroup

	reader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			redisLogger.Printf("连接断开: %v", err)
			return
		}
		redisLogger.Printf("收到数据: %s", strings.TrimSpace(message))

		parts := strings.Split(strings.TrimSpace(message), " ")
		if len(parts) == 0 {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			switch strings.ToUpper(parts[0]) {
			case "PING":
				if !authenticated {
					redisLogger.Printf("认证需求: 需要认证才能执行PING命令")
					_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				redisLogger.Printf("收到PING命令")
				_, err := conn.Write([]byte("+PONG\r\n"))
				if err != nil {
					redisLogger.Printf("写入响应出错: %v", err)
				}
			case "AUTH":
				redisLogger.Printf("收到AUTH命令")
				if len(parts) < 2 {
					redisLogger.Printf("无效认证参数")
					_, err := conn.Write([]byte("-ERR Invalid AUTH parameters\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				var providedPassword string
				var providedUsername string
				if len(parts) == 2 {
					providedPassword = parts[1]
				} else if len(parts) == 3 {
					providedUsername = parts[1]
					providedPassword = parts[2]
				} else {
					redisLogger.Printf("无效认证参数")
					_, err := conn.Write([]byte("-ERR Invalid AUTH parameters\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}

				if (cfg.Redis.Username == "" || providedUsername == cfg.Redis.Username) && providedPassword == cfg.Redis.Password {
					redisLogger.Printf("认证成功")
					_, err := conn.Write([]byte("+OK Authentication successful.\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					authenticated = true
				} else {
					redisLogger.Printf("认证失败")
					_, err := conn.Write([]byte("-ERR invalid password\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
				}
			case "SET":
				if !authenticated {
					redisLogger.Printf("认证需求: 需要认证才能执行SET命令")
					_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				if len(parts) < 3 {
					redisLogger.Printf("SET命令参数无效")
					_, err := conn.Write([]byte("-ERR wrong number of arguments for 'set'\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				key := parts[1]
				value := strings.Join(parts[2:], " ")
				redisLogger.Printf("执行SET命令，键: %s, 值: %s", key, value)
				_, err := conn.Write([]byte("+OK\r\n"))
				if err != nil {
					redisLogger.Printf("写入响应出错: %v", err)
				}
			case "GET":
				if !authenticated {
					redisLogger.Printf("认证需求: 需要认证才能执行GET命令")
					_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				if len(parts) < 2 {
					redisLogger.Printf("GET命令参数无效")
					_, err := conn.Write([]byte("-ERR wrong number of arguments for 'get'\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				key := parts[1]
				redisLogger.Printf("执行GET命令，键: %s", key)
				_, err := conn.Write([]byte("$0\r\n\r\n"))
				if err != nil {
					redisLogger.Printf("写入响应出错: %v", err)
				}
			default:
				if !authenticated {
					redisLogger.Printf("认证需求: 需要认证才能执行该命令")
					_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
					if err != nil {
						redisLogger.Printf("写入响应出错: %v", err)
					}
					return
				}
				redisLogger.Printf("收到未知命令: %s", parts[0])
				_, err := conn.Write([]byte("-ERR Unknown command\r\n"))
				if err != nil {
					redisLogger.Printf("写入响应出错: %v", err)
				}
			}
		}()

		if err != nil {
			return
		}
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	}

	wg.Wait()
	conn.Close()
}

// GetServiceName返回服务名称
func (m *SimRedisService) GetServiceName() string {
	return "Redis"
}
