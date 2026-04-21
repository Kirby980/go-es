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

// Debug ...
func (NopLogger) Debug(msg string, keysAndValues ...any) {}

// Info ...
func (NopLogger) Info(msg string, keysAndValues ...any) {}

// Warn ...
func (NopLogger) Warn(msg string, keysAndValues ...any) {}

// Error ...
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

// Debug ...
func (l *ZapLogger) Debug(msg string, keysAndValues ...any) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Info ...
func (l *ZapLogger) Info(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warn ...
func (l *ZapLogger) Warn(msg string, keysAndValues ...any) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Error ...
func (l *ZapLogger) Error(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}
