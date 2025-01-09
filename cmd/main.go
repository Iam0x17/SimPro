package main

import (
	"SimPro/common"
	"SimPro/services"
	"fmt"
	"github.com/fatih/color"
	"github.com/projectdiscovery/goflags"
	"github.com/sirupsen/logrus"
	"log"
	"os"
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

	CommandInit()
	fmt.Printf("silent: %s \n", Targets)
	lines := []string{
		` █████                   █████                       `,
		`░░███                   ░░███                        `,
		` ░███        █████ ████ ███████   ████████   ██████  `,
		` ░███       ░░███ ░███ ░░░███░   ░░███░░███ ░░░░░███ `,
		` ░███        ░███ ░███   ░███     ░███ ░░░   ███████ `,
		` ░███      █ ░███ ░███   ░███ ███ ░███      ███░░███ `,
		` ███████████ ░░████████  ░░█████  █████    ░░████████`,
		`░░░░░░░░░░░   ░░░░░░░░    ░░░░░  ░░░░░      ░░░░░░░░ `,
	}

	for i, line := range lines {
		coloredLine := color.New(color.Attribute(i + 100)).Sprint(line)
		fmt.Println(coloredLine)
	}
	//fmt.Println("" +
	//	" █████                   █████                       \n" +
	//	"░░███                   ░░███                        \n" +
	//	" ░███        █████ ████ ███████   ████████   ██████  \n" +
	//	" ░███       ░░███ ░███ ░░░███░   ░░███░░███ ░░░░░███ \n" +
	//	" ░███        ░███ ░███   ░███     ░███ ░░░   ███████ \n" +
	//	" ░███      █ ░███ ░███   ░███ ███ ░███      ███░░███ \n" +
	//	" ███████████ ░░████████  ░░█████  █████    ░░████████\n" +
	//	"░░░░░░░░░░░   ░░░░░░░░    ░░░░░  ░░░░░      ░░░░░░░░ ")

	//startSSH := flag.Bool("ssh", false, "Start the SSH service")
	//startFTP := flag.Bool("ftp", false, "Start the FTP service")
	//startRedis := flag.Bool("redis", false, "Start the Redis service")
	//startTelnet := flag.Bool("telnet", false, "Start the Telnet service")
	//
	//flag.Parse()
	//
	//manager := &services.ServiceManager{}
	//
	//addServiceIfFlagSet(manager, startSSH, &ssh.SimSSHService{})
	//addServiceIfFlagSet(manager, startFTP, &ftp.SimFTPService{})
	//addServiceIfFlagSet(manager, startRedis, &redis.SimRedisService{})
	//addServiceIfFlagSet(manager, startTelnet, &telnet.SimTelnetService{})
	//
	//ctx := context.Background()
	//manager.StartAllServices(ctx)
	//
	//// 防止主程序退出，维持服务运行
	//select {}
}
