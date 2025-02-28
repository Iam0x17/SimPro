# SimPro - 轻量级协议模拟器
![](https://img.shields.io/badge/Go-1.20%2B-blue)  


![](https://img.shields.io/badge/License-Apache%202.0-green)

# <font style="color:rgb(204, 204, 204);">📋</font><font style="color:rgb(204, 204, 204);"> </font>项目介绍
在进行安全验证工作时，我们需要对各类服务开展暴力破解测试，这就要求能够模拟多种服务。在搜寻开源项目的时候发现了fapro。fapro很强大，但由于其闭源，无法进行二次开发来满足需求，因此有了SimPro这个项目。

SimPro 是用Go语言构建的协议模拟器，采用模块化架构实现多协议(FTP/SSH/DB等)服务。可用于安全验证和简单充当蜜罐服务。

# 🚀 功能介绍
通过命令行参数/HTTP接口控制服务启停

支持SSH、Redis、PostgreSQL、MySQL、Telnet、FTP等多种协议服务

结构化日志记录

## 命令参数
```php
Application Options:
  /s, /services:  要启动的服务，以逗号分隔
  /c, /config:    配置文件路径
  /l, /log:       日志文件路径
  /v, /verbose    详细打印caller

Help Options:
  /?              Show this help message
  /h, /help       Show this help message
```

## 多协议支持
| 现阶段支持的协议 |
| --- |
| FTP |
| SSH |
| Telnet |
| MySQL |
| Redis |
| PostgreSQL |


# 🔍功能演示
启动SSH服务

`go run main.go /s ssh`

ssh远程连接，前2次输入错误密码，第3次输入正确密码模拟登录，输入whoami及ls命令的返回结果

![](/docs/ssh_connect.png)

日志结构化记录了远程登录ip、登录账户密码、执行命令等信息

![](/docs/ssh_server.png)

# <font style="color:rgb(204, 204, 204);">📊</font>后续计划
1.增加各种服务

2.增加C2模拟回连服务，模拟C2服务器连接implant，让implant正常上线和心跳

