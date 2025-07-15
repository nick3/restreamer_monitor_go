# Restreamer Monitor Go

[English](#english) | [中文](#中文)

## 中文

Restreamer Monitor Go 是一个用 Go 语言开发的多平台直播间监测与转播工具。

### 功能特性

- **多平台支持**: 目前支持 Bilibili 直播平台，架构设计支持扩展其他平台
- **实时监控**: 实时监控直播间状态，检测开播和下播
- **直播转播**: 支持将直播流转播到多个目标平台（RTMP/RTMPS）
- **多目标推流**: 同时推流到多个目标地址，支持不同质量设置
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

#### 转播功能

转播功能可以将源直播流转发到多个目标平台：

```bash
# 启动转播服务
./RestreamerMonitor relay -c config.json -v

# 使用特定质量设置
./RestreamerMonitor relay -c config.json -q 720p -v
```

支持的质量设置：
- `best`: 最佳质量（原始质量）
- `720p`: 720p 分辨率，2000k 码率
- `480p`: 480p 分辨率，1000k 码率
- `worst`: 最低质量，500k 码率

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
  "interval": "30s",
  "verbose": false
}
```

配置说明：
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