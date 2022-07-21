package logger

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/pipego/scheduler/config"
)

type User struct {
	Name string
}

type Log struct {
	Level string `json:"level"`
	Ts    string `json:"ts"`
	Msg   string `json:"msg"`
	Name  string `json:"name"`
	User  User   `json:"user"`
	Func  string `json:"func"`
	File  string `json:"file"`
	Line  int    `json:"line"`
}

var (
	cfg = config.Config{
		Spec: config.Spec{
			Logger: config.Logger{
				CallerSkip:   2,
				FileCompress: false,
				FileName:     "test.log",
				LogLevel:     "debug",
				MaxAge:       1,
				MaxBackups:   60,
				MaxSize:      100,
			},
		},
	}
)

var buf = &User{Name: "name"}

func readFile(name string) (Log, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return Log{}, err
	}

	var buf Log
	err = json.Unmarshal(content, &buf)

	return buf, err
}

func TestDebug(t *testing.T) {
	ctx := context.Background()

	l := logger{
		cfg: &Config{
			Config: cfg,
		},
	}

	_ = l.Init(ctx)

	l.Debug("debug", zap.String("name", buf.Name), zap.Any("user", buf))
	buf, err := readFile(cfg.Spec.Logger.FileName)
	_ = os.Remove(cfg.Spec.Logger.FileName)

	assert.Equal(t, nil, err)
	assert.Equal(t, "debug", buf.Msg)

	_ = l.Deinit(ctx)
}

func TestInfo(t *testing.T) {
	ctx := context.Background()

	l := logger{
		cfg: &Config{
			Config: cfg,
		},
	}

	_ = l.Init(ctx)

	l.Info("info", zap.String("name", buf.Name), zap.Any("user", buf))
	buf, err := readFile(cfg.Spec.Logger.FileName)
	_ = os.Remove(cfg.Spec.Logger.FileName)

	assert.Equal(t, nil, err)
	assert.Equal(t, "info", buf.Msg)

	_ = l.Deinit(ctx)
}

func TestWarn(t *testing.T) {
	ctx := context.Background()

	l := logger{
		cfg: &Config{
			Config: cfg,
		},
	}

	_ = l.Init(ctx)

	l.Warn("warn", zap.String("name", buf.Name), zap.Any("user", buf))
	buf, err := readFile(cfg.Spec.Logger.FileName)
	_ = os.Remove(cfg.Spec.Logger.FileName)

	assert.Equal(t, nil, err)
	assert.Equal(t, "warn", buf.Msg)

	_ = l.Deinit(ctx)
}

func TestError(t *testing.T) {
	ctx := context.Background()

	l := logger{
		cfg: &Config{
			Config: cfg,
		},
	}

	_ = l.Init(ctx)

	l.Error("error", zap.String("name", buf.Name), zap.Any("user", buf))
	buf, err := readFile(cfg.Spec.Logger.FileName)
	_ = os.Remove(cfg.Spec.Logger.FileName)

	assert.Equal(t, nil, err)
	assert.Equal(t, "error", buf.Msg)

	_ = l.Deinit(ctx)
}
