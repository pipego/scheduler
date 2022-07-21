package logger

import (
	"context"
	"os"
	"path"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/pipego/scheduler/config"
)

var (
	LogLevel = map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}
)

type Logger interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Debug(string, ...zap.Field)
	Info(string, ...zap.Field)
	Warn(string, ...zap.Field)
	Error(string, ...zap.Field)
}

type Config struct {
	Config config.Config
}

type logger struct {
	cfg *Config
	zap *zap.Logger
}

func New(_ context.Context, cfg *Config) Logger {
	return &logger{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (l *logger) Init(_ context.Context) error {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	enc := zapcore.NewJSONEncoder(cfg)

	writer := &lumberjack.Logger{
		Compress:   l.cfg.Config.Spec.Logger.FileCompress,
		Filename:   l.cfg.Config.Spec.Logger.FileName,
		MaxAge:     int(l.cfg.Config.Spec.Logger.MaxAge),
		MaxBackups: int(l.cfg.Config.Spec.Logger.MaxBackups),
		MaxSize:    int(l.cfg.Config.Spec.Logger.MaxSize),
	}

	core := zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), zapcore.WarnLevel),
		zapcore.NewCore(enc, zapcore.AddSync(writer), LogLevel[l.cfg.Config.Spec.Logger.LogLevel]),
	)

	l.zap = zap.New(core)

	return nil
}

func (l *logger) Deinit(_ context.Context) error {
	return nil
}

func (l *logger) Debug(message string, fields ...zap.Field) {
	info := l.callerFields()
	fields = append(fields, info...)
	l.zap.Debug(message, fields...)
}

func (l *logger) Info(message string, fields ...zap.Field) {
	info := l.callerFields()
	fields = append(fields, info...)
	l.zap.Info(message, fields...)
}

func (l *logger) Warn(message string, fields ...zap.Field) {
	info := l.callerFields()
	fields = append(fields, info...)
	l.zap.Warn(message, fields...)
}

func (l *logger) Error(message string, fields ...zap.Field) {
	info := l.callerFields()
	fields = append(fields, info...)
	l.zap.Error(message, fields...)
}

func (l *logger) callerFields() (fields []zap.Field) {
	pc, file, line, ok := runtime.Caller(int(l.cfg.Config.Spec.Logger.CallerSkip))
	if !ok {
		return nil
	}

	funcName := runtime.FuncForPC(pc).Name()
	funcName = path.Base(funcName)

	fields = append(fields, zap.String("func", funcName), zap.String("file", file), zap.Int("line", line))

	return
}
