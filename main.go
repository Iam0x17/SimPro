package main

import (
	"SimPro/api/http"
	"SimPro/common"
	"SimPro/config"
	"SimPro/services"
	"SimPro/services/ftp"
	"SimPro/services/mysql"
	"SimPro/services/postgres"
	"SimPro/services/redis"
	"SimPro/services/ssh"
	"SimPro/services/telnet"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Services   []string `long:"services" short:"s" description:"要启动的服务，以逗号分隔"`
	ConfigPath string   `long:"config" short:"c" description:"配置文件路径"`
	LogPath    string   `long:"log" short:"l" description:"日志文件路径"`
	Verbose    bool     `long:"verbose" short:"v" description:"详细打印caller"`
}

//go:embed  assets/*
var assetsFs embed.FS

func main() {

	config.GetEmbed(assetsFs)
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化全局日志器
	common.InitLogger(opts.Verbose, opts.LogPath)
	defer common.SyncLogger()

	var servicesToStart []string
	for _, s := range opts.Services {
		parts := strings.Split(s, ",")
		for _, p := range parts {
			servicesToStart = append(servicesToStart, strings.TrimSpace(p))
		}
	}

	err = http.StartHttpService()
	if err != nil {
		panic(err)
	}

	cfg, err := config.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	manager := services.NewServiceManager(cfg)
	manager.AddService(&ssh.SimSSHService{})
	manager.AddService(&redis.SimRedisService{})
	manager.AddService(&postgres.SimPostgresService{})
	manager.AddService(&mysql.SimMySqlService{})
	manager.AddService(&telnet.SimTelnetService{})
	manager.AddService(&ftp.SimFTPService{})

	for _, s := range servicesToStart {
		err := manager.StartServiceByName(s)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出信号
	<-sigChan
}
