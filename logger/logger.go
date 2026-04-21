package logger

import "go.uber.org/zap"

// Logger 日志接口，允许用户注入自定义日志实现
// 方法签名采用 SugaredLogger 风格（key-value 对），兼容 zap、logrus、slog 等主流库
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// -----------------------------------------------------------------------
// NopLogger：静默实现，不输出任何内容
// -----------------------------------------------------------------------

// NopLogger 空日志实现，用于完全禁用日志输出
type NopLogger struct{}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (NopLogger) Debug(msg string, keysAndValues ...any) {}

// Info 记录一条 Info 级别的日志信息。
func (NopLogger) Info(msg string, keysAndValues ...any) {}

// Warn 记录一条 Warn (警告) 级别的日志信息。
func (NopLogger) Warn(msg string, keysAndValues ...any) {}

// Error 记录一条 Error 级别的日志信息。
func (NopLogger) Error(msg string, keysAndValues ...any) {}

// -----------------------------------------------------------------------
// ZapLogger：基于 go.uber.org/zap 的实现（默认）
// -----------------------------------------------------------------------

// ZapLogger 封装 zap.SugaredLogger，实现 Logger 接口
type ZapLogger struct {
	sugar *zap.SugaredLogger
}

// NewZapLogger 用已有的 *zap.Logger 创建 ZapLogger
// 示例：
//
//	z, _ := zap.NewProduction()
//	l := logger.NewZapLogger(z)
func NewZapLogger(z *zap.Logger) *ZapLogger {
	return &ZapLogger{sugar: z.Sugar()}
}

// NewDefaultLogger 创建默认的生产级 zap 日志（JSON 格式，输出到 stderr，Info 级别及以上）
// 用于 client.New() 未显式指定 Logger 时的默认值
func NewDefaultLogger() (*ZapLogger, error) {
	z, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return NewZapLogger(z), nil
}

// NewDevelopmentLogger 创建适合开发调试的 zap 日志（人类可读格式，Debug 级别及以上）
// 示例：
//
//	l, _ := logger.NewDevelopmentLogger()
//	client, _ := client.New(config.WithLogger(l))
func NewDevelopmentLogger() (*ZapLogger, error) {
	z, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return NewZapLogger(z), nil
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (l *ZapLogger) Debug(msg string, keysAndValues ...any) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Info 记录一条 Info 级别的日志信息。
func (l *ZapLogger) Info(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warn 记录一条 Warn (警告) 级别的日志信息。
func (l *ZapLogger) Warn(msg string, keysAndValues ...any) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Error 记录一条 Error 级别的日志信息。
func (l *ZapLogger) Error(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}
