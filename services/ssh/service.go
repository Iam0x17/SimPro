package ssh

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"unicode/utf8"

	"SimPro/common"
	"SimPro/config"

	"golang.org/x/crypto/ssh"
)

// SimSSHService 实现服务接口
type SimSSHService struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func (s *SimSSHService) Start(cfg *config.Config) error {

	// 配置 SSH 服务器
	sshCfg := setupSSHServerConfig(cfg)

	// 加载或创建私钥
	private, err := loadOrCreatePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to load or create private key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return fmt.Errorf("failed to create signer: %v", err)
		//common.Logger.
	}
	sshCfg.AddHostKey(signer)

	// 监听端口
	listener, err := net.Listen("tcp", ":"+cfg.SSH.Port)
	if err != nil {
		return fmt.Errorf("listening port failed: %v", err)
	}
	s.listener = listener
	common.Logger.Info(common.EventStartService, zap.String("protocol", "ssh"), zap.String("info", fmt.Sprintf("SSH service is listening on port %s", cfg.SSH.Port)))
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					//sshLogger.Printf("接受连接失败: %v", err)
					return
				}
				continue
			}
			// 建立 SSH 连接
			sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshCfg)
			if err != nil {
				//sshLogger.Printf("建立 SSH 连接失败: %v，错误: %v", conn.RemoteAddr(), err)
				common.Logger.Info(common.EventNewConnection, zap.String("protocol", "ssh"), zap.String("info", fmt.Sprintf("Failed to establish SSH connection: %v，error: %v", conn.RemoteAddr(), err)))
				continue
			}
			defer sshConn.Close()

			//sshLogger.Printf("新的 SSH 服务连接来自 %s，版本 %v，用户: %v", sshConn.RemoteAddr(), sshConn.ClientVersion(), sshConn.User())
			common.Logger.Info(common.EventNewConnection, zap.String("protocol", "ssh"),
				zap.String("info", fmt.Sprintf("Version: %s，User: %v", string(sshConn.ClientVersion()), sshConn.User())),
				zap.String("local", sshConn.LocalAddr().String()),
				zap.String("remote", sshConn.RemoteAddr().String()))

			// 处理 SSH 通道请求
			go ssh.DiscardRequests(reqs)

			// 处理 SSH 通道
			s.wg.Add(1)
			go handleConnection(sshConn, chans, cfg)
		}
	}()

	return nil
}

func (s *SimSSHService) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
	}
	s.wg.Wait()
	common.Logger.Info(common.EventStopService, zap.String("protocol", "ssh"), zap.String("info", "SSH service has stopped"))
	return nil
}

func (s *SimSSHService) GetName() string {
	return "ssh"
}

func setupSSHServerConfig(cfg *config.Config) *ssh.ServerConfig {
	sshCfg := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			//sshLogger.Printf("收到来自 %s 的密码认证尝试，用户: %s，密码: %s", conn.RemoteAddr(), conn.User(), string(password))

			if cfg.SSH.User == conn.User() && cfg.SSH.Pass == string(password) {
				common.Logger.Info(common.EventAccountLogin,
					zap.String("protocol", "ssh"),
					zap.String("info", "Login succeeded"),
					zap.String("account", conn.User()),
					zap.String("password", string(password)),
					zap.String("local", conn.LocalAddr().String()),
					zap.String("remote", conn.RemoteAddr().String()))
				return &ssh.Permissions{}, nil
			}
			common.Logger.Info(common.EventAccountLogin,
				zap.String("protocol", "ssh"),
				zap.String("info", "Login failed"),
				zap.String("account", conn.User()),
				zap.String("password", string(password)),
				zap.String("local", conn.LocalAddr().String()),
				zap.String("remote", conn.RemoteAddr().String()))
			return nil, fmt.Errorf("password error")
		},
		NoClientAuth:  false,
		ServerVersion: "SSH-2.0-SimSSH",
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			//sshLogger.Printf("收到来自 %s 的公钥认证尝试，公钥: %v", conn.RemoteAddr(), key)
			return nil, fmt.Errorf("public key authentication failed")
		},
	}
	return sshCfg
}

func handleConnection(sshConn *ssh.ServerConn, chans <-chan ssh.NewChannel, cfg *config.Config) {
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "Unknown channel type")
			continue
		}
		//sshLogger.Printf("接受来自 %v 的新通道，类型 :%v", sshConn.RemoteAddr(), newChannel.ChannelType())
		channel, requests, err := newChannel.Accept()
		if err != nil {
			//sshLogger.Printf("Unable to accept channel: %v", err)
			continue
		}
		term := make(chan string, 1)
		go handleChannel(channel, requests, sshConn, term, cfg)
	}
}

func handleChannel(channel ssh.Channel, requests <-chan *ssh.Request, sshConn *ssh.ServerConn, term chan<- string, cfg *config.Config) {
	defer func(channel ssh.Channel) {
		err := channel.Close()
		if err != nil {

		}
	}(channel)

	commands := cfg.SSH.Commands
	for req := range requests {
		switch req.Type {
		case "env":
			err := req.Reply(true, nil)
			if err != nil {
				return
			}
			//sshLogger.Printf("收到来自 %v 的 env 请求 %v", sshConn.RemoteAddr(), req.Type)
		case "exec":
			//sshLogger.Printf("收到来自 %s 的 Exec 请求 %v", sshConn.RemoteAddr(), string(req.Payload[4:]))
			handleExec(req, channel, sshConn, commands)
		case "shell":
			//sshLogger.Printf("收到来自 %s 的 Shell 请求", sshConn.RemoteAddr())
			err := req.Reply(true, nil)
			if err != nil {
				return
			}
			handleShell(channel, sshConn, commands)
		case "pty-req":
			terminal := string(req.Payload[4 : len(req.Payload)-4])
			if !utf8.ValidString(terminal) {
				//sshLogger.Printf("来自 %s 的无效终端类型，使用默认终端 ", sshConn.RemoteAddr())
				terminal = "xterm-256color"
			}
			term <- terminal
			err := req.Reply(true, nil)
			if err != nil {
				return
			}
			//sshLogger.Printf("收到来自 %v 的 pty 请求 %v，终端:%v", sshConn.RemoteAddr(), req.Type, terminal)
		default:
			//sshLogger.Printf("未知请求类型 %s 来自 %v", req.Type, sshConn.RemoteAddr())
			err := req.Reply(false, nil)
			if err != nil {
				return
			}
		}
	}
}

func handleExec(req *ssh.Request, channel ssh.Channel, sshConn *ssh.ServerConn, commands map[string]string) {
	command := string(req.Payload[4:])
	//sshLogger.Printf("收到来自 %s 的 Exec 请求 %v", sshConn.RemoteAddr(), command)
	output := executeCommand(command, commands)
	//sshLogger.Printf("向客户端写入: %v，来自 %v", output, sshConn.RemoteAddr())
	_, err := channel.Write([]byte(output))
	if err != nil {
		return
	}
	common.Logger.Info(common.EventReplyCommand,
		zap.String("protocol", "ssh"),
		zap.String("account", sshConn.User()),
		zap.String("info", output),
		zap.String("local", sshConn.LocalAddr().String()),
		zap.String("remote", sshConn.RemoteAddr().String()))
	//sshLogger.Printf("发送 shell 请求 来自 %v", sshConn.RemoteAddr())
	err = req.Reply(true, nil)
	if err != nil {
		return
	}
	channel.Close()
}

func handleShell(channel ssh.Channel, sshConn *ssh.ServerConn, commands map[string]string) {
	common.Logger.Info(common.EventStartShell,
		zap.String("protocol", "ssh"),
		zap.String("account", sshConn.User()),
		zap.String("info", "start shell"),
		zap.String("local", sshConn.LocalAddr().String()),
		zap.String("remote", sshConn.RemoteAddr().String()))
	//sshLogger.Printf("开始处理来自 %v 的 shell", sshConn.RemoteAddr())
	defer common.Logger.Info(common.EventStopShell,
		zap.String("protocol", "ssh"),
		zap.String("account", sshConn.User()),
		zap.String("info", "stop shell"),
		zap.String("local", sshConn.LocalAddr().String()),
		zap.String("remote", sshConn.RemoteAddr().String()))

	// 发送欢迎消息
	sendWelcomeMessage(channel, sshConn)

	// 模拟 shell 提示符
	prompt := "root@simpro-ssh:~# "
	channel.Write([]byte(prompt))
	//sshLogger.Printf("向客户端写入提示符，来自 %v", sshConn.RemoteAddr())

	var command string
	buf := make([]byte, 1)

	for {
		n, err := channel.Read(buf)
		if err != nil {
			//if err != io.EOF {
			//	sshLogger.Printf("读取错误: %v", err)
			//} else {
			//	sshLogger.Printf("读取到 EOF，来自 :%v", sshConn.RemoteAddr())
			//}
			break
		}

		if n > 0 {
			b := buf[0]

			// 回显用户输入
			channel.Write([]byte{b})

			if b == 13 {
				// 处理命令之前去除首尾空格
				command = strings.TrimSpace(command)
				//sshLogger.Printf("收到来自 %s 的命令: %s", sshConn.RemoteAddr(), command)
				common.Logger.Info(common.EventExecuteCommand,
					zap.String("protocol", "ssh"),
					zap.String("account", sshConn.User()),
					zap.String("info", strings.ReplaceAll(command, "\\u0008", "")),
					zap.String("local", sshConn.LocalAddr().String()),
					zap.String("remote", sshConn.RemoteAddr().String()))
				if command == "exit" {
					channel.Write([]byte("\r\n"))
					channel.Close()
					return
				}

				output := executeCommand(command, commands)
				channel.Write([]byte("\r\n" + output))
				//sshLogger.Printf("向客户端写入 : %v，来自 %v", output, sshConn.RemoteAddr())
				common.Logger.Info(common.EventReplyCommand,
					zap.String("protocol", "ssh"),
					zap.String("account", sshConn.User()),
					zap.String("info", output),
					zap.String("local", sshConn.LocalAddr().String()),
					zap.String("remote", sshConn.RemoteAddr().String()))
				channel.Write([]byte("\r\n"))
				channel.Write([]byte(prompt))
				//sshLogger.Printf("向客户端写入提示符，来自 %v", sshConn.RemoteAddr())
				command = ""
			} else {
				command += string(b)
			}
		}
	}
}

func sendWelcomeMessage(channel ssh.Channel, sshConn *ssh.ServerConn) {
	welcomeMsg := "\r\nWelcome to Simulating SSH Server!\r\n"
	channel.Write([]byte(welcomeMsg))
	//sshLogger.Printf("向客户端写入欢迎消息，来自 %v", sshConn.RemoteAddr())
	channel.SendRequest("shell", true, nil)
	//sshLogger.Printf("发送 shell 请求 来自 %v", sshConn.RemoteAddr())
}

func executeCommand(command string, commands map[string]string) string {
	if val, ok := commands[command]; ok {
		return val
	}
	return "bash: " + command + " command not found\n"
}

// loadOrCreatePrivateKey 加载或创建私键
func loadOrCreatePrivateKey() (crypto.PrivateKey, error) {
	var privateKey crypto.PrivateKey
	data, err := config.AssetsFs.ReadFile(config.SshPrivateKey)
	//data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block != nil {
		switch block.Type {
		case "RSA PRIVATE KEY":
			privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		}
	}
	//if err != nil || privateKey == nil {
	//	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	//	if err != nil {
	//		return nil, err
	//	}
	//	var pemKey = &pem.Block{
	//		Type:  "RSA PRIVATE KEY",
	//		Bytes: x509.MarshalPKCS1PrivateKey(privateKey.(*rsa.PrivateKey)),
	//	}
	//	var pemData = pem.EncodeToMemory(pemKey)
	//	config.AssetsFs.
	//	//err = os.WriteFile(filename, pemData, 0600)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	return privateKey, nil
}
