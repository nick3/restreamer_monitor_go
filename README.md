# Restreamer Monitor Go

[English](#english) | [中文](#中文)

## 中文

Restreamer Monitor Go 是一个用 Go 语言开发的多平台直播间监测与转播工具。

### 功能特性

- **多平台支持**: 目前支持 Bilibili 直播平台，架构设计支持扩展其他平台
- **实时监控**: 实时监控直播间状态，检测开播和下播
- **命令行界面**: 基于 Cobra 的友好命令行界面
- **配置灵活**: 支持 JSON 配置文件和命令行参数
- **高性能**: 使用 Go 协程实现高并发监控
- **安全可靠**: 完善的错误处理和输入验证

### 安装

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

# 转播直播流（开发中）
./RestreamerMonitor relay -c config.json
```

#### 配置文件

创建 `config.json` 配置文件：

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

#### 命令参数

**monitor 命令:**
- `-c, --config`: 指定配置文件路径（默认: ../config.json）
- `-i, --interval`: 监控检查间隔（默认: 30s）
- `-v, --verbose`: 启用详细日志输出

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