package main

import (
	"SimPro/api/http"
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
	"github.com/jessevdk/go-flags"
	"log"
	"strings"
)

type Options struct {
	Services   []string `long:"services" short:"s" description:"要启动的服务，以逗号分隔"`
	ConfigPath string   `long:"config" short:"c" description:"配置文件路径"`
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

	var servicesToStart []string
	for _, s := range opts.Services {
		parts := strings.Split(s, ",")
		for _, p := range parts {
			servicesToStart = append(servicesToStart, strings.TrimSpace(p))
		}
	}

	//fmt.Printf("Targets: %s \n", strings.Join(Targets, ", "))
	err = http.StartHttpService()
	if err != nil {
		panic(err)
	}
	fmt.Println("路径:" + opts.ConfigPath)
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

	//fmt.Println("元素:", opts.Services)

	for _, s := range servicesToStart {
		err := manager.StartServiceByName(s)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	select {}
}
