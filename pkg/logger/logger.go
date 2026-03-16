package logger

import (
	"os"

	"go.uber.org/zap"			// Zap日志库
	"go.uber.org/zap/zapcore"
)

// 全局日志实例，供封装的方法调用
var log *zap.Logger

// Init 初始化日志配置：根据运行模式设置日志级别
func Init(mode string) {
	var level zapcore.Level
	if mode == "release" {
		level = zapcore.InfoLevel	// 生产环境只输出Info 及以上级别（Info/Warn/Error/Fatal）
	} else {
		level = zapcore.DebugLevel  // 非生产环境（如 debug）：输出 Debug 及以上所有级别
	}

	// 日志编码器（JSON格式）
	encoderCfg := zap.NewProductionEncoderConfig() 		// 基础生产环境编码器配置
	encoderCfg.TimeKey = "time"                   		// 日志中时间字段的 key 为 "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder 	// 时间格式：ISO8601（如 2026-03-14T12:34:56.789Z）

	// 创建 Zap 核心组件
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg), // 编码器：输出 JSON 格式日志
		zapcore.AddSync(os.Stdout),         // 输出目标：标准输出（控制台）
		level,                              // 日志级别阈值
	)

	// 创建全局日志实例
	log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// 封装常用日志方法，简化业务代码调用
func Info(msg string, fields ...zap.Field)  { log.Info(msg, fields...) }  // 信息级日志
func Error(msg string, fields ...zap.Field) { log.Error(msg, fields...) } // 错误级日志
func Warn(msg string, fields ...zap.Field)  { log.Warn(msg, fields...) }  // 警告级日志
func Debug(msg string, fields ...zap.Field) { log.Debug(msg, fields...) } // 调试级日志
func Fatal(msg string, fields ...zap.Field) { log.Fatal(msg, fields...) } // 致命级日志（会触发 os.Exit(1)）
func Sync()                                 { _ = log.Sync() } 			  // 同步日志缓冲区（确保日志写入完成）
