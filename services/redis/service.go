package redis

import (
	"SimPro/common"
	"bufio"
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"SimPro/config"
)

type SimRedisService struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func (s *SimRedisService) GetName() string {
	return "redis"
}

func (s *SimRedisService) Start(cfg *config.Config) error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+cfg.Redis.Port)
	if err != nil {
		return err
	}
	common.Logger.Info(common.EventStartService, zap.String("protocol", "redis"), zap.String("info", fmt.Sprintf("Redis service is listening on port %s", cfg.Redis.Port)))
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {

				if err != net.ErrClosed {
					return
				}
				continue
			}
			s.wg.Add(1)
			go handleConnection(conn, cfg)
		}
	}()
	return nil
}

func (s *SimRedisService) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
	}
	s.wg.Wait()
	common.Logger.Info(common.EventStopService, zap.String("protocol", "redis"), zap.String("info", "Redis service has stopped"))
	return nil
}

func scanRedisProtocol(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if data[0] == '*' { // 判断是否为数组
		// 处理数组
		parts := bytes.Split(data, []byte("\r\n"))
		if len(parts) < 2 {
			return 0, nil, nil
		}

		arrayLen, err := strconv.Atoi(string(parts[0][1:]))
		if err != nil {
			return 0, nil, fmt.Errorf("invalid array length: %w", err)
		}

		expectedLen := arrayLen*2 + 1
		if len(parts) < expectedLen {
			return 0, nil, nil // Not enough data
		}

		return len(bytes.Join(parts[:expectedLen], []byte("\r\n"))) + 2, bytes.Join(parts[:expectedLen], []byte("\r\n")), nil

	} else if data[0] == '$' { // 判断是否为字符串
		parts := bytes.Split(data, []byte("\r\n"))
		if len(parts) < 2 {
			return 0, nil, nil
		}

		strLen, err := strconv.Atoi(string(parts[0][1:]))
		if err != nil {
			return 0, nil, fmt.Errorf("invalid string length: %w", err)
		}

		expectedLen := 2
		if len(parts) < expectedLen {
			return 0, nil, nil // Not enough data
		}
		totalLen := len(parts[0]) + 2 + strLen + 2
		if len(data) < totalLen {
			return 0, nil, nil
		}
		return totalLen, data[:totalLen-2], nil
	} else if i := bytes.Index(data, []byte("\r\n")); i >= 0 {
		return i + 2, data[:i], nil // 如果是简单字符串，直接分割
	}

	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func parseRedisCommand(message string) ([]string, error) {
	if message == "" {
		return nil, nil
	}

	if message[0] == '*' {
		parts := strings.Split(message, "\r\n")

		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid array format: %s", message)
		}

		arrayLen, err := strconv.Atoi(string(parts[0][1:]))
		if err != nil {
			return nil, fmt.Errorf("invalid array length: %w", err)
		}

		if len(parts) < arrayLen*2+1 {
			return nil, fmt.Errorf("invalid array data: %s", message)
		}
		command := make([]string, arrayLen)
		for i := 1; i <= arrayLen; i++ {
			command[i-1] = parts[i*2]
		}
		return command, nil
	} else if message[0] == '$' {
		parts := strings.Split(message, "\r\n")

		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid string format: %s", message)
		}

		strLen, err := strconv.Atoi(string(parts[0][1:]))
		if err != nil {
			return nil, fmt.Errorf("invalid string length: %w", err)
		}

		if strLen != len(parts[1]) {
			return nil, fmt.Errorf("invalid string length: %w", err)
		}

		return []string{parts[1]}, nil

	} else {
		return strings.Split(message, " "), nil
	}

}

func handleConnection(conn net.Conn, cfg *config.Config) {
	var authenticated bool

	scanner := bufio.NewScanner(conn)
	scanner.Split(scanRedisProtocol)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	common.Logger.Info(common.EventNewConnection, zap.String("protocol", "redis"),
		zap.String("info", "New redis connect"),
		zap.String("local", conn.LocalAddr().String()),
		zap.String("remote", conn.RemoteAddr().String()))

	for scanner.Scan() {
		message := scanner.Text()
		//----
		//redisLogger.Printf("收到数据: %s", message)
		//fmt.Printf("收到数据: %s", message)

		parts, err := parseRedisCommand(message)
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
			_, err := conn.Write([]byte("-ERR Invalid command format\r\n"))
			if err != nil {
				common.Logger.Error("Error", zap.Error(err))
				//redisLogger.Printf("写入响应出错: %v", err)
			}
			continue
		}
		if len(parts) == 0 {
			continue
		}
		//common.Logger.Info(common.EventExecuteCommand, zap.String("protocol", "redis"),
		//	zap.String("info", parts[0]),
		//	zap.String("local", conn.LocalAddr().String()),
		//	zap.String("remote", conn.RemoteAddr().String()))
		switch strings.ToUpper(parts[0]) {
		case "PING":
			handlePingCommand(conn, authenticated)
		case "AUTH":
			handleAuthCommand(conn, parts, cfg)
			authenticated = handleAuthentication(parts, cfg)
		case "SET":
			handleSetCommand(conn, parts, authenticated)
		case "GET":
			handleGetCommand(conn, parts, authenticated)
		default:
			handleUnknownCommand(conn, authenticated)
		}

		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	}
	if err := scanner.Err(); err != nil {
		//common.Logger.Error("Error", zap.Error(err))
		common.Logger.Info(common.EventCloseConnection, zap.String("protocol", "redis"),
			zap.String("info", "close redis connect"),
			zap.String("local", conn.LocalAddr().String()),
			zap.String("remote", conn.RemoteAddr().String()))
	}
}

func handlePingCommand(conn net.Conn, authenticated bool) {
	if !authenticated {
		//redisLogger.Printf("认证需求: 需要认证才能执行PING命令")
		_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	//redisLogger.Printf("收到PING命令")
	_, err := conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		common.Logger.Error("Error", zap.Error(err))
	}
}

func handleAuthCommand(conn net.Conn, parts []string, cfg *config.Config) {
	//redisLogger.Printf("收到AUTH命令")
	if len(parts) < 2 {
		//redisLogger.Printf("无效认证参数")
		_, err := conn.Write([]byte("-ERR Invalid AUTH parameters\r\n"))
		if err != nil {
			//redisLogger.Printf("写入响应出错: %v", err)
			common.Logger.Error("Error", zap.Error(err))
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
		//redisLogger.Printf("无效认证参数")
		_, err := conn.Write([]byte("-ERR Invalid AUTH parameters\r\n"))
		if err != nil {
			//redisLogger.Printf("写入响应出错: %v", err)
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	if (cfg.Redis.User == "" || providedUsername == cfg.Redis.User) && providedPassword == cfg.Redis.Pass {
		//redisLogger.Printf("认证成功")
		//_, err := conn.Write([]byte("+OK Authentication successful.\r\n"))
		common.Logger.Info(common.EventAccountLogin,
			zap.String("protocol", "redis"),
			zap.String("info", "Login succeeded"),
			zap.String("account", providedUsername),
			zap.String("password", providedPassword),
			zap.String("local", conn.LocalAddr().String()),
			zap.String("remote", conn.RemoteAddr().String()))
		_, err := conn.Write([]byte("+OK\r\n"))
		if err != nil {
			//redisLogger.Printf("写入响应出错: %v", err)
			common.Logger.Error("Error", zap.Error(err))
		}
	} else {
		//redisLogger.Printf("认证失败")
		common.Logger.Info(common.EventAccountLogin,
			zap.String("protocol", "redis"),
			zap.String("info", "Login failed"),
			zap.String("account", providedUsername),
			zap.String("password", providedPassword),
			zap.String("local", conn.LocalAddr().String()),
			zap.String("remote", conn.RemoteAddr().String()))
		_, err := conn.Write([]byte("-ERR invalid password\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
	}
}

func handleAuthentication(parts []string, cfg *config.Config) bool {
	var providedPassword string
	var providedUsername string
	if len(parts) == 2 {
		providedPassword = parts[1]
	} else if len(parts) == 3 {
		providedUsername = parts[1]
		providedPassword = parts[2]
	}
	return (cfg.Redis.User == "" || providedUsername == cfg.Redis.User) && providedPassword == cfg.Redis.Pass
}

func handleSetCommand(conn net.Conn, parts []string, authenticated bool) {
	if !authenticated {
		//redisLogger.Printf("认证需求: 需要认证才能执行SET命令")
		_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	if len(parts) < 3 {
		//redisLogger.Printf("SET命令参数无效")
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'set'\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	//key := parts[1]
	//value := strings.Join(parts[2:], " ")
	//----
	//redisLogger.Printf("执行SET命令，键: %s, 值: %s", key, value)
	//fmt.Printf("执行SET命令，键: %s, 值: %s", key, value)
	_, err := conn.Write([]byte("+OK\r\n"))
	if err != nil {
		common.Logger.Error("Error", zap.Error(err))
	}
}

func handleGetCommand(conn net.Conn, parts []string, authenticated bool) {
	if !authenticated {
		//redisLogger.Printf("认证需求: 需要认证才能执行GET命令")
		_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	if len(parts) < 2 {
		//redisLogger.Printf("GET命令参数无效")
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'get'\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	//key := parts[1]
	//redisLogger.Printf("执行GET命令，键: %s", key)
	//----
	//fmt.Printf("执行GET命令，键: %s", key)
	_, err := conn.Write([]byte("$0\r\n\r\n"))
	if err != nil {
		common.Logger.Error("Error", zap.Error(err))
	}
}

func handleUnknownCommand(conn net.Conn, authenticated bool) {
	if !authenticated {
		//redisLogger.Printf("认证需求: 需要认证才能执行该命令")
		_, err := conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
		if err != nil {
			common.Logger.Error("Error", zap.Error(err))
		}
		return
	}
	//redisLogger.Printf("收到未知命令")
	_, err := conn.Write([]byte("-ERR Unknown command\r\n"))
	if err != nil {
		common.Logger.Error("Error", zap.Error(err))
	}
}
