# SimPro - 轻量级协议模拟器
![](https://img.shields.io/badge/Go-1.20%2B-blue)  


![](https://img.shields.io/badge/License-Apache%202.0-green)

## 📋 项目介绍
在进行安全验证工作时，我们需要对各类服务开展暴力破解测试，这就要求能够模拟多种服务。SimPro是一个用Go语言构建的轻量级协议模拟器，采用模块化架构实现多协议(FTP/SSH/DB等)服务模拟。可用于安全验证和简单充当蜜罐服务。

### 主要特性
- 🚀 支持多种常见协议服务模拟
- 📝 详细的结构化日志记录
- 🌐 Web管理界面
- ⚙️ 灵活的配置系统
- 🔌 模块化设计，易于扩展

## 快速开始

### 环境要求
- Go 1.20+
- 操作系统：Windows/Linux/MacOS

### 安装
1. 克隆仓库
```bash
git clone https://github.com/Iam0x17/SimPro.git
cd SimPro
```

2. 编译
```bash
build.bat
```
## 命令参数
```php
Application Options:
  /s, /services:  要启动的服务，以逗号分隔
  /c, /config:    配置文件路径
  /l, /log:       日志文件路径
  /v, /verbose    详细打印caller
  /w, /web        启动web服务器
  /p, /port:      web服务器端口 (default: 8080)

Help Options:
  /?              Show this help message
  /h, /help       Show this help message
```

### 基本使用
1. 启动单个服务
```bash
SimPro /s ssh
```

2. 启动多个服务
```bash
SimPro /s ssh,ftp,mysql
```

3. 启动Web管理界面
```bash
SimPro /w /p 8000
```

## 📚 详细文档
- [更新日志](CHANGELOG.md)

## 支持的协议
| 协议 | 默认端口 | 状态 |
|------|---------|------|
| SSH | 22 | ✅ |
| FTP | 21 | ✅ |
| Telnet | 23 | ✅ |
| MySQL | 3306 | ✅ |
| Redis | 6379 | ✅ |
| PostgreSQL | 5432 | ✅ |

## 配置示例
```yaml
ssh:
  port: 2222
  user: root
  pass: 123456
  commands:
    "whoami": "root\n"
    "uname -a": "Linux fake-honeypot 5.15.0-75-generic #82-Ubuntu SMP Mon Jun 19 14:18:11 UTC 2023 x86_64 x86_64 x86_64 GNU/Linux\n"
ftp:
  port: 2121
  user: root
  pass: 123456
redis:
  port: 6379
  user: root
  pass: 123456
telnet:
  port: 2323
  user: root
  pass: 123456
postgres:
  port: 5432
  user: postgres
  pass: 123456
mysql:
  port: 3306
  user: root
  pass: 123456
```

## 🔍 功能演示
### SSH服务模拟

ssh远程连接，前2次输入错误密码，第3次输入正确密码模拟登录，输入whoami及ls命令的返回结果
![](/docs/ssh_connect.png)


日志结构化记录了远程登录ip、登录账户密码、执行命令等信息
![](/docs/ssh_server.png)

### Web管理界面

服务管理
![](/docs/web_manger.png)

服务配置
![](/docs/web_config.png)

## 📊 开发计划
- [√] 增加更多协议支持
  - [ ] HTTP/HTTPS
  - [ ] SMTP
  - ...
- [?] 增加C2模拟回连服务
- [?] 增加数据统计功能
- [√] 完善Web管理界面

