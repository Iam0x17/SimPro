package main

import (
	"ProtoSimService/services"
	"context"
	"flag"

	"ProtoSimService/services/ftp"
	"ProtoSimService/services/redis"
	"ProtoSimService/services/ssh"
)

func main() {
	startSSH := flag.Bool("ssh", false, "Start the SSH service")
	startFTP := flag.Bool("ftp", false, "Start the FTP service")
	startRedis := flag.Bool("redis", false, "Start the Redis service")

	flag.Parse()

	manager := &services.ServiceManager{}

	if *startSSH {
		sshService := &ssh.MockSSHService{}
		manager.AddService(sshService)
	}

	if *startFTP {
		ftpService := &ftp.MockFTPService{}
		manager.AddService(ftpService)
	}

	if *startRedis {
		redisService := &redis.MockRedisService{}
		manager.AddService(redisService)
	}
	ctx := context.Background()
	manager.StartAllServices(ctx)

	// 防止主程序退出，维持服务运行
	select {}
}
