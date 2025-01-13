package main

import (
	"SimPro/common"
	"SimPro/services"
	"SimPro/services/ftp"
	"SimPro/services/mysql"
	"SimPro/services/redis"
	"SimPro/services/ssh"
	"SimPro/services/telnet"
	"context"
	"flag"
	"fmt"
	"github.com/projectdiscovery/goflags"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var Targets goflags.StringSlice

func addServiceIfFlagSet(manager *services.ServiceManager, flagValue *bool, service common.SimService) {
	if *flagValue {
		err := manager.AddService(service)
		if err != nil {
			logrus.Errorf("Failed to add service: %v", err)
			os.Exit(1)
		}
	}
}

func CommandInit() {
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`欢迎使用协议模拟服务`)

	flagSet.CreateGroup("services", "Services",
		flagSet.StringSliceVarP(&Targets, "services", "s", nil, "启动的服务，可以为多个，多个用英文逗号分割。", goflags.FileCommaSeparatedStringSliceOptions),
	)

	if err := flagSet.Parse(); err != nil {
		log.Fatalf("Could not parse flags: %s\n", err)
	}
}

func main() {

	//CommandInit()
	//fmt.Printf("silent: %s \n", Targets)
	//fmt.Printf("Targets: %s \n", strings.Join(Targets, ", "))

	startSSH := flag.Bool("ssh", false, "Start the SSH service")
	startFTP := flag.Bool("ftp", false, "Start the FTP service")
	startRedis := flag.Bool("redis", false, "Start the Redis service")
	startTelnet := flag.Bool("telnet", false, "Start the Telnet service")
	startMysql := flag.Bool("mysql", false, "Start the Mysql service")

	flag.Parse()

	manager := &services.ServiceManager{}

	addServiceIfFlagSet(manager, startSSH, &ssh.SimSSHService{})
	addServiceIfFlagSet(manager, startFTP, &ftp.SimFTPService{})
	addServiceIfFlagSet(manager, startRedis, &redis.SimRedisService{})
	addServiceIfFlagSet(manager, startTelnet, &telnet.SimTelnetService{})
	addServiceIfFlagSet(manager, startMysql, &mysql.SimMySqlService{})

	ctx := context.Background()
	manager.StartAllServices(ctx)

	// 信号处理
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("Received signal:", sig)
		// 在这里添加清理服务、关闭资源的逻辑
		os.Exit(0)
	}()

	// 等待信号
	select {}
}

//func main() {
//	startSSH := flag.Bool("ssh", false, "Start the SSH service")
//	startFTP := flag.Bool("ftp", false, "Start the FTP service")
//	startRedis := flag.Bool("redis", false, "Start the Redis service")
//
//	flag.Parse()
//
//	manager := &services.ServiceManager{}
//
//	if *startSSH {
//		sshService := &ssh.SimSSHService{}
//		manager.AddService(sshService)
//	}
//
//	if *startFTP {
//		ftpService := &ftp.SimFTPService{}
//		manager.AddService(ftpService)
//	}
//
//	if *startRedis {
//		redisService := &redis.SimRedisService{}
//		manager.AddService(redisService)
//	}
//	ctx := context.Background()
//	manager.StartAllServices(ctx)
//
//	// 防止主程序退出，维持服务运行
//	select {}
//}
