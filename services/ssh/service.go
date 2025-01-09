package ssh

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"SimPro/common"
	"SimPro/config"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var sshLogger *logrus.Logger

func init() {
	sshLogger = common.SetupServiceLogger("SSH", true)
}

// MockSSHService 实现通用的MockService接口
type SimSSHService struct{}

func (m *SimSSHService) NeedsListener() bool {
	return false
}

func (m *SimSSHService) ServeWithListener(ctx context.Context, listener net.Listener) {

}

// Serve 处理SSH连接相关逻辑
func (m *SimSSHService) Serve(ctx context.Context, conn net.Conn) {
	cfg, err := config.LoadConfig()
	if err != nil {
		sshLogger.Fatalf("加载配置失败: %v", err)
	}
	commands := cfg.SSH.Commands

	// 配置SSH服务器
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			sshLogger.Printf("收到来自 %s 的密码认证尝试，用户: %s，密码: %s", conn.RemoteAddr(), conn.User(), string(password))
			if pass, ok := cfg.SSH.ValidUsers[conn.User()]; ok && pass == string(password) {
				return &ssh.Permissions{}, nil
			}
			return nil, fmt.Errorf("密码不允许")
		},
		NoClientAuth:  false,
		ServerVersion: "SSH-2.0-SimSSH",
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			sshLogger.Printf("收到来自 %s 的公钥认证尝试，公钥: %v", conn.RemoteAddr(), key)
			return nil, fmt.Errorf("公钥认证不允许")
		},
	}

	// 加载或创建私钥
	private, err := loadOrCreatePrivateKey("host.key")
	if err != nil {
		sshLogger.Fatal("加载或创建私钥失败: ", err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		sshLogger.Fatalf("创建签名器失败: %v", err)
	}
	config.AddHostKey(signer)

	// 建立SSH连接
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		sshLogger.Printf("建立SSH连接失败: %v，错误: %v", conn.RemoteAddr(), err)
		return
	}
	sshLogger.Printf("新的SSH连接来自 %s，版本 %v，用户: %v", sshConn.RemoteAddr(), sshConn.ClientVersion(), sshConn.User())

	defer sshConn.Close()

	// 处理SSH通道请求
	go ssh.DiscardRequests(reqs)

	// 处理SSH通道
	wg := sync.WaitGroup{}
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "未知通道类型")
			continue
		}
		sshLogger.Printf("接受来自 %v 的新通道，类型 :%v", conn.RemoteAddr(), newChannel.ChannelType())
		channel, requests, err := newChannel.Accept()
		if err != nil {
			sshLogger.Printf("无法接受通道: %v", err)
			continue
		}
		wg.Add(1)
		term := make(chan string, 1)
		go func(in <-chan *ssh.Request) {
			defer wg.Done()
			var terminal string
			for req := range in {
				switch req.Type {
				case "env":
					req.Reply(true, nil)
					sshLogger.Printf("收到来自 %v 的env请求 %v", sshConn.RemoteAddr(), req.Type)
				case "exec":
					sshLogger.Printf("收到来自 %s 的Exec请求 %v", sshConn.RemoteAddr(), string(req.Payload[4:]))
					handleExec(req, channel, sshConn, commands)
				case "shell":
					sshLogger.Printf("收到来自 %s 的Shell请求", sshConn.RemoteAddr())
					req.Reply(true, nil)
					handleShell(channel, sshConn, terminal, commands)
				case "pty-req":
					terminal = string(req.Payload[4 : len(req.Payload)-4])
					if !utf8.ValidString(terminal) {
						sshLogger.Printf("来自 %s 的无效终端类型，使用默认终端 ", sshConn.RemoteAddr())
						terminal = "xterm-256color"
					}
					term <- terminal
					req.Reply(true, nil)
					sshLogger.Printf("收到来自 %v 的pty请求 %v，终端:%v", sshConn.RemoteAddr(), req.Type, terminal)
				default:
					sshLogger.Printf("未知请求类型 %s 来自 %v", req.Type, sshConn.RemoteAddr())
					req.Reply(false, nil)
				}
			}
		}(requests)
		<-term
	}
	wg.Wait()
	conn.Close()
}

func handleExec(req *ssh.Request, channel ssh.Channel, sshConn *ssh.ServerConn, commands map[string]string) {
	command := string(req.Payload[4:])
	sshLogger.Printf("收到来自 %s 的Exec请求 %v", sshConn.RemoteAddr(), command)
	var output string
	if val, ok := commands[command]; ok {
		output = val
	} else {
		output = ""
	}
	sshLogger.Printf("向客户端写入: %v，来自 %v", output, sshConn.RemoteAddr())
	channel.Write([]byte(output))
	sshLogger.Printf("发送shell请求 来自 %v", sshConn.RemoteAddr())
	channel.SendRequest("shell", true, nil)
	req.Reply(true, nil)
	channel.Close()
}

func handleShell(channel ssh.Channel, sshConn *ssh.ServerConn, terminal string, commands map[string]string) {
	sshLogger.Printf("开始处理来自 %v 的shell", sshConn.RemoteAddr())
	defer sshLogger.Printf("结束处理来自 %v 的shell", sshConn.RemoteAddr())

	// 发送欢迎消息
	channel.Write([]byte("\r\n欢迎来到模拟SSH服务器!\r\n"))
	sshLogger.Printf("向客户端写入欢迎消息，来自 %v", sshConn.RemoteAddr())

	// 发送一次shell请求
	sshLogger.Printf("发送shell请求 来自 %v", sshConn.RemoteAddr())
	channel.SendRequest("shell", true, nil)

	if terminal == "" {
		terminal = "xterm-256color"
		fmt.Println("使用默认终端")
	} else {
		fmt.Printf("客户端终端 : %v \n", terminal)
	}

	// 模拟shell提示符
	prompt := "root@mock-ssh:~# "
	channel.Write([]byte(prompt))
	sshLogger.Printf("向客户端写入提示符，来自 %v", sshConn.RemoteAddr())

	var command string
	buf := make([]byte, 1)

	for {
		n, err := channel.Read(buf)
		if err != nil {
			if err != io.EOF {
				sshLogger.Printf("读取错误: %v", err)
			} else {
				sshLogger.Printf("读取到EOF，来自 :%v", sshConn.RemoteAddr())
			}
			break
		}

		if n > 0 {
			b := buf[0]

			// 回显用户输入
			channel.Write([]byte{b})

			if b == 13 {
				// 处理命令之前去除首尾空格
				command = strings.TrimSpace(command)
				sshLogger.Printf("收到来自 %s 的命令: %s", sshConn.RemoteAddr(), command)

				if command == "exit" {
					channel.Write([]byte("\r\n"))
					channel.Close()
					return
				}

				var output string
				if val, ok := commands[command]; ok {
					output = val
				} else {
					output = "bash: " + command + ": 命令未找到\n"
				}

				channel.Write([]byte("\r\n" + output))
				sshLogger.Printf("向客户端写入 : %v，来自 %v", output, sshConn.RemoteAddr())
				channel.Write([]byte("\r\n"))
				channel.Write([]byte(prompt))
				sshLogger.Printf("向客户端写入提示符，来自 %v", sshConn.RemoteAddr())
				command = ""
			} else {
				command += string(b)
			}
		}
	}
}

// loadOrCreatePrivateKey 加载或创建私键
func loadOrCreatePrivateKey(filename string) (crypto.PrivateKey, error) {
	var privateKey crypto.PrivateKey
	data, err := os.ReadFile(filename)
	if err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			switch block.Type {
			case "RSA PRIVATE KEY":
				privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			}
		}
	}
	if err != nil || privateKey == nil {
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		var pemKey = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey.(*rsa.PrivateKey)),
		}
		var pemData = pem.EncodeToMemory(pemKey)
		err = os.WriteFile(filename, pemData, 0600)
		if err != nil {
			return nil, err
		}
	}
	return privateKey, nil
}

// GetServiceName 返回服务名称
func (m *SimSSHService) GetServiceName() string {
	return "SSH"
}
