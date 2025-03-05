package app

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
	Web        bool     `long:"web" short:"w" description:"启动web服务器"`
	WebPort    string   `long:"port" short:"p" description:"web服务器端口" default:"8080"`
}

type App struct {
	opts     *Options
	manager  *services.ServiceManager
	assetsFs embed.FS
}

func NewApp(assetsFs embed.FS) *App {
	return &App{
		assetsFs: assetsFs,
	}
}

func (a *App) parseCommandLineArgs() error {
	a.opts = &Options{}
	parser := flags.NewParser(a.opts, flags.HelpFlag)

	_, err := parser.Parse()
	if err != nil {
		parser.WriteHelp(os.Stderr)
		return err
	}

	return nil
}

func (a *App) initializeServices() error {
	cfg, err := config.LoadConfig(a.opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	a.manager = services.NewServiceManager(cfg)
	a.manager.AddService(&ssh.SimSSHService{})
	a.manager.AddService(&redis.SimRedisService{})
	a.manager.AddService(&postgres.SimPostgresService{})
	a.manager.AddService(&mysql.SimMySqlService{})
	a.manager.AddService(&telnet.SimTelnetService{})
	a.manager.AddService(&ftp.SimFTPService{})

	return nil
}

func (a *App) startRequestedServices() error {
	for _, s := range a.opts.Services {
		parts := strings.Split(s, ",")
		for _, p := range parts {
			serviceName := strings.TrimSpace(p)
			err := a.manager.StartServiceByName(serviceName)
			if err != nil {
				return fmt.Errorf("启动服务 %s 失败: %v", serviceName, err)
			}
		}
	}
	return nil
}

func (a *App) setupSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	return sigChan
}

func (a *App) Run() error {
	config.GetEmbed(a.assetsFs)

	// 解析命令行参数
	if err := a.parseCommandLineArgs(); err != nil {
		return err
	}

	// 初始化全局日志器
	common.InitLogger(a.opts.Verbose, a.opts.LogPath)
	defer common.SyncLogger()

	// 启动web服务器（如果指定）
	if a.opts.Web {
		if err := http.StartHttpService(a.opts.WebPort, a.assetsFs); err != nil {
			return err
		}
	}

	// 如果没有指定任何服务且没有启动web服务器，显示帮助信息并退出
	if len(a.opts.Services) == 0 && !a.opts.Web {
		flags.NewParser(&Options{}, flags.HelpFlag).WriteHelp(os.Stderr)
		return fmt.Errorf("未指定任何服务或web服务器")
	}

	// 初始化服务管理器
	if err := a.initializeServices(); err != nil {
		return err
	}

	// 启动请求的服务
	if err := a.startRequestedServices(); err != nil {
		return err
	}

	// 等待退出信号
	<-a.setupSignalHandler()
	return nil
}
