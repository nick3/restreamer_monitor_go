# Restreamer Monitor Go

[English](#english) | [中文](#中文)

## 中文

Restreamer Monitor Go 是一个用 Go 语言开发的多平台直播间监测与转播工具。

### 功能特性

- **多平台支持**: 目前支持 Bilibili 直播平台，架构设计支持扩展其他平台
- **实时监控**: 实时监控直播间状态，检测开播和下播
- **直播转播**: 支持将直播流转播到多个目标平台（RTMP/RTMPS）
- **多目标推流**: 同时推流到多个目标地址，支持不同质量设置
- **Telegram Bot集成**: 完整的Telegram Bot支持，实时通知和远程控制
- **智能通知系统**: 支持系统、监控、转播和错误等多种通知类型
- **远程控制**: 通过Telegram Bot远程控制服务启停和状态查询
- **命令行界面**: 基于 Cobra 的友好命令行界面
- **配置灵活**: 支持 JSON 配置文件和命令行参数
- **高性能**: 使用 Go 协程实现高并发处理
- **FFmpeg 集成**: 利用 FFmpeg 进行高效的流媒体处理
- **安全可靠**: 完善的错误处理和输入验证

### 安装

#### 系统要求

- Go 1.21 或更高版本
- FFmpeg（用于转播功能）

```bash
# 在 Ubuntu/Debian 上安装 FFmpeg
sudo apt update
sudo apt install ffmpeg

# 在 macOS 上安装 FFmpeg
brew install ffmpeg

# 在 CentOS/RHEL 上安装 FFmpeg
sudo yum install epel-release
sudo yum install ffmpeg
```

#### 从源码编译

```bash
git clone https://github.com/nick3/restreamer_monitor_go.git
cd restreamer_monitor_go
make build
```

#### 跨平台编译

```bash
# 编译所有支持的平台
make build-all

# 编译特定平台
GOOS=linux GOARCH=amd64 go build -o bin/RestreamerMonitor_linux ./main/main.go
```

#### Docker 部署

推荐使用 Docker 部署，支持 amd64 和 arm64 架构。

**拉取镜像：**

```bash
# 拉取最新版本
docker pull ghcr.io/nick3/restreamer_monitor_go:latest

# 拉取指定版本
docker pull ghcr.io/nick3/restreamer_monitor_go:v1.0.0
```

**准备配置文件：**

在本地创建配置文件 `config.json`（参考下方配置文件示例）。

**运行容器：**

```bash
# 监控模式
docker run -d \
  --name restreamer-monitor \
  -v $(pwd)/config.json:/app/config/config.json:ro \
  ghcr.io/nick3/restreamer_monitor_go:latest \
  monitor -c /app/config/config.json -v

# 转播模式
docker run -d \
  --name restreamer-relay \
  -v $(pwd)/config.json:/app/config/config.json:ro \
  ghcr.io/nick3/restreamer_monitor_go:latest \
  relay -c /app/config/config.json -v
```

**使用 Docker Compose：**

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  restreamer-monitor:
    image: ghcr.io/nick3/restreamer_monitor_go:latest
    container_name: restreamer-monitor
    restart: unless-stopped
    volumes:
      - ./config.json:/app/config/config.json:ro
    command: ["monitor", "-c", "/app/config/config.json", "-v"]
```

启动服务：

```bash
docker-compose up -d
```

**本地构建镜像：**

```bash
# 构建镜像
docker build -t restreamer-monitor .

# 运行本地构建的镜像
docker run -d \
  --name restreamer-monitor \
  -v $(pwd)/config.json:/app/config/config.json:ro \
  restreamer-monitor \
  monitor -c /app/config/config.json -v
```

### 使用方法

#### 基本命令

```bash
# 显示帮助信息
./RestreamerMonitor --help

# 监控直播间状态
./RestreamerMonitor monitor -c config.json -i 30s -v

# 转播直播流
./RestreamerMonitor relay -c config.json -v

# 查看版本信息
./RestreamerMonitor --version
```

#### Telegram Bot 功能

项目完整集成了 Telegram Bot，支持：

```bash
# 启用 Telegram Bot 通知
./RestreamerMonitor monitor -c config.json

# 启用 Telegram Bot 转播控制
./RestreamerMonitor relay -c config.json
```

**Bot 命令列表：**
- `/start` - 显示欢迎信息
- `/help` - 显示帮助信息
- `/status` - 查看系统运行状态
- `/rooms` - 查看监控房间状态
- `/relays` - 查看转播状态
- `/stop [service]` - 停止指定服务（monitor/relay）
- `/restart [service]` - 重启指定服务（monitor/relay/system）

**通知类型：**
- 🖥️ 系统事件：启动、停止、重启
- 👁️ 监控事件：开播、下播状态变化
- 🔄 转播事件：转播启动、停止、错误
- ❌ 错误事件：系统错误和异常

**安全特性：**
- 管理员权限控制
- 命令启用/禁用配置
- 多聊天室通知支持

#### 配置文件

创建 `config.json` 配置文件：

```json
{
  "rooms": [
    {
      "platform": "bilibili",
      "room_id": "123456",
      "enabled": true
    }
  ],
  "relays": [
    {
      "name": "bilibili-to-multiple",
      "source": {
        "platform": "bilibili",
        "room_id": "123456"
      },
      "destinations": [
        {
          "name": "youtube",
          "url": "rtmp://a.rtmp.youtube.com/live2/YOUR_YOUTUBE_STREAM_KEY",
          "protocol": "rtmp",
          "options": {
            "bufsize": "3000k",
            "maxrate": "3000k"
          }
        },
        {
          "name": "twitch",
          "url": "rtmp://live.twitch.tv/live/YOUR_TWITCH_STREAM_KEY",
          "protocol": "rtmp",
          "options": {
            "bufsize": "6000k",
            "maxrate": "6000k"
          }
        }
      ],
      "enabled": true,
      "quality": "720p"
    }
  ],
  "telegram": {
    "enabled": true,
    "bot_token": "YOUR_BOT_TOKEN_HERE",
    "chat_ids": [123456789, -1001234567890],
    "admin_ids": [123456789],
    "enabled_commands": ["start", "help", "status", "rooms", "relays", "stop", "restart"],
    "notifications": {
      "system_events": true,
      "monitor_events": true,
      "relay_events": true,
      "error_events": true
    }
  },
  "interval": "30s",
  "verbose": false
}
```

**Telegram 配置说明：**
- `bot_token`: Telegram Bot Token（从 @BotFather 获取）
- `chat_ids`: 接收通知的聊天ID列表（个人聊天或群组）
- `admin_ids`: 管理员用户ID列表（可执行控制命令）
- `enabled_commands`: 启用的Bot命令列表
- `notifications`: 各类通知的开关设置

**其他配置说明：**
- `rooms`: 监控的直播间列表
- `relays`: 转播配置列表
- `source`: 源直播间信息
- `destinations`: 目标推流地址列表
- `quality`: 流质量设置
- `options`: FFmpeg 额外参数

#### 命令参数

**monitor 命令:**
- `-c, --config`: 指定配置文件路径（默认: ../config.json）
- `-i, --interval`: 监控检查间隔（默认: 30s）
- `-v, --verbose`: 启用详细日志输出

**relay 命令:**
- `-c, --config`: 指定配置文件路径（默认: ../config.json）
- `-v, --verbose`: 启用详细日志输出
- `-q, --quality`: 指定流质量（best, worst, 720p, 480p）

### API 文档

#### 核心接口

```go
// StreamSource 定义直播源接口
type StreamSource interface {
    GetStatus() bool                    // 获取直播状态
    GetRoomInfo() models.RoomInfo      // 获取房间信息  
    GetPlayURL() string                // 获取播放URL
    StartMsgListener()                 // 开始消息监听
    CloseMsgListener()                 // 关闭消息监听
}

// BilibiliService Bilibili API 服务
type BilibiliService struct {
    RoomId string
    Client *resty.Client
}
```

#### 主要方法

```go
// 创建 Bilibili 服务实例
func NewBilibiliService(roomId string) (*BilibiliService, error)

// 获取真实房间号
func (b *BilibiliService) GetBilibiliRealRoomId() (string, error)

// 获取直播状态
func (b *BilibiliService) GetBilibiliLiveStatus() (bool, error)

// 获取直播流URL
func (b *BilibiliService) GetBilibiliLiveRealURL(realRoomId string) ([]string, error)
```

### 开发

#### 项目结构

```
restreamer_monitor_go/
├── cli/            # 命令行界面
├── main/           # 主程序入口
├── models/         # 数据模型
├── monitor/        # 监控逻辑
├── service/        # 第三方服务接口
├── bin/            # 编译输出
├── Makefile        # 构建脚本
└── go.mod          # Go 模块定义
```

#### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 代码质量

项目遵循以下最佳实践：

- **安全性**: 输入验证、错误处理、避免敏感信息泄露
- **性能**: HTTP 连接复用、重试机制、超时控制
- **可维护性**: 清晰的代码结构、完善的文档、全面的测试
- **可扩展性**: 接口设计支持多平台扩展

### 贡献

欢迎提交 Issue 和 Pull Request！

### 许可证

本项目采用 MIT 许可证。

---

## English

Restreamer Monitor Go is a multi-platform live streaming monitoring and restreaming tool developed in Go.

### Features

- **Multi-platform Support**: Currently supports Bilibili platform with extensible architecture for other platforms
- **Real-time Monitoring**: Real-time monitoring of live room status, detecting stream start/stop events
- **Command Line Interface**: User-friendly CLI based on Cobra framework
- **Flexible Configuration**: Support for JSON config files and command line parameters
- **High Performance**: Concurrent monitoring using Go routines
- **Security & Reliability**: Comprehensive error handling and input validation

### Installation

#### Build from Source

```bash
git clone https://github.com/nick3/restreamer_monitor_go.git
cd restreamer_monitor_go
make build
```

#### Cross-platform Build

```bash
# Build for all supported platforms
make build-all

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/RestreamerMonitor_linux ./main/main.go
```

#### Docker Deployment

Docker deployment is recommended, supporting both amd64 and arm64 architectures.

**Pull the image:**

```bash
# Pull latest version
docker pull ghcr.io/nick3/restreamer_monitor_go:latest

# Pull specific version
docker pull ghcr.io/nick3/restreamer_monitor_go:v1.0.0
```

**Prepare configuration file:**

Create a local `config.json` configuration file (refer to the configuration file example below).

**Run the container:**

```bash
# Monitor mode
docker run -d \
  --name restreamer-monitor \
  -v $(pwd)/config.json:/app/config/config.json:ro \
  ghcr.io/nick3/restreamer_monitor_go:latest \
  monitor -c /app/config/config.json -v

# Relay mode
docker run -d \
  --name restreamer-relay \
  -v $(pwd)/config.json:/app/config/config.json:ro \
  ghcr.io/nick3/restreamer_monitor_go:latest \
  relay -c /app/config/config.json -v
```

**Using Docker Compose:**

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  restreamer-monitor:
    image: ghcr.io/nick3/restreamer_monitor_go:latest
    container_name: restreamer-monitor
    restart: unless-stopped
    volumes:
      - ./config.json:/app/config/config.json:ro
    command: ["monitor", "-c", "/app/config/config.json", "-v"]
```

Start the service:

```bash
docker-compose up -d
```

**Build image locally:**

```bash
# Build image
docker build -t restreamer-monitor .

# Run locally built image
docker run -d \
  --name restreamer-monitor \
  -v $(pwd)/config.json:/app/config/config.json:ro \
  restreamer-monitor \
  monitor -c /app/config/config.json -v
```

### Usage

#### Basic Commands

```bash
# Show help
./RestreamerMonitor --help

# Monitor live room status
./RestreamerMonitor monitor -c config.json -i 30s -v

# Relay live streams (under development)
./RestreamerMonitor relay -c config.json
```

#### Configuration File

Create a `config.json` configuration file:

```json
{
  "rooms": [
    {
      "platform": "bilibili",
      "room_id": "123456", 
      "enabled": true
    },
    {
      "platform": "bilibili",
      "room_id": "789012",
      "enabled": false
    }
  ],
  "interval": "30s",
  "verbose": false
}
```

#### Command Options

**monitor command:**
- `-c, --config`: Specify config file path (default: ../config.json)
- `-i, --interval`: Monitoring check interval (default: 30s)
- `-v, --verbose`: Enable verbose logging

### API Documentation

#### Core Interfaces

```go
// StreamSource defines the live stream source interface
type StreamSource interface {
    GetStatus() bool                    // Get live status
    GetRoomInfo() models.RoomInfo      // Get room information
    GetPlayURL() string                // Get play URL
    StartMsgListener()                 // Start message listener
    CloseMsgListener()                 // Close message listener
}

// BilibiliService Bilibili API service
type BilibiliService struct {
    RoomId string
    Client *resty.Client
}
```

### Development

#### Project Structure

```
restreamer_monitor_go/
├── cli/            # Command line interface
├── main/           # Main program entry
├── models/         # Data models
├── monitor/        # Monitoring logic
├── service/        # Third-party service interfaces
├── bin/            # Build output
├── Makefile        # Build scripts
└── go.mod          # Go module definition
```

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Contributing

Issues and Pull Requests are welcome!

### License

This project is licensed under the MIT License.