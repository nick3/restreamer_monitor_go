package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局logger实例
	Logger *logrus.Logger
	// Entry 带默认字段的entry
	Entry *logrus.Entry
)

// Config 日志配置
type Config struct {
	// Level 日志等级 (trace, debug, info, warn, error, fatal, panic)
	Level string `json:"level"`
	// LogFile 日志文件路径
	LogFile string `json:"log_file"`
	// Console 是否输出到控制台
	Console bool `json:"console"`
	// MaxSize 日志文件最大尺寸(MB)
	MaxSize int `json:"max_size"`
	// MaxBackups 最大备份文件数
	MaxBackups int `json:"max_backups"`
	// MaxAge 最大保留天数
	MaxAge int `json:"max_age"`
	// Compress 是否压缩/归档旧日志文件
	Compress bool `json:"compress"`
	// AppName 应用名称
	AppName string `json:"app_name"`
	// AppVersion 版本号
	AppVersion string `json:"app_version"`
	// Environment 运行环境 (development/production)
	Environment string `json:"environment"`
	// ReportCaller 是否启用日志调用位置报告
	ReportCaller bool `json:"report_caller"`
}

// DefaultConfig 返回默认日志配置
func DefaultConfig() Config {
	return Config{
		Level:        "info",
		LogFile:      "logs/restreamer-monitor.log",
		Console:      true,
		MaxSize:      100,    // 100MB
		MaxBackups:   10,     // 保留10个旧文件
		MaxAge:       30,     // 保留30天
		Compress:     true,   // 压缩旧日志
		AppName:      "RestreamerMonitor",
		AppVersion:   "1.0.0",
		Environment:  "production",
		ReportCaller: false,  // 生产环境关闭以提高性能
	}
}

// InitLogger 初始化日志系统
func InitLogger(cfg *Config) error {
	Logger = logrus.New()

	// 设置日志等级
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		// 如果解析失败，默认使用info级别
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置日志格式 - 文本格式，带时间戳和颜色
	Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableColors:    false,
		DisableTimestamp: false,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	})

	// 设置调用位置报告
	Logger.SetReportCaller(cfg.ReportCaller)

	// 配置输出目标
	var writers []io.Writer

	// 控制台输出
	if cfg.Console {
		writers = append(writers, os.Stdout)
	}

	// 文件输出（配置了日志文件路径时）
	if cfg.LogFile != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 配置lumberjack日志轮转
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}

		writers = append(writers, lumberjackLogger)
	}

	// 设置多输出目标
	if len(writers) > 0 {
		Logger.SetOutput(io.MultiWriter(writers...))
	}

	// 创建带默认字段的entry
	Entry = Logger.WithFields(logrus.Fields{
		"app":         cfg.AppName,
		"version":     cfg.AppVersion,
		"environment": cfg.Environment,
	})

	return nil
}

// InitCompatLogger 初始化兼容模式，重定向标准库log
func InitCompatLogger() {
	if Logger == nil {
		return
	}
	// 重定向标准库log到logrus
	log.SetOutput(Logger.Writer())
	log.SetFlags(0) // 移除标准库的时间戳和前缀
}

// GetLogger 获取带上下文的logger
func GetLogger(fields map[string]interface{}) *logrus.Entry {
	if Entry == nil {
		return logrus.WithFields(fields)
	}

	if len(fields) == 0 {
		return Entry
	}
	return Entry.WithFields(fields)
}
