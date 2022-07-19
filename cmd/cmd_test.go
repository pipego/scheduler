package cmd

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/pipego/scheduler/config"
)

func testInitConfig() *config.Config {
	cfg := config.New()

	fi, _ := os.Open("../test/config/config.yml")

	defer func() {
		_ = fi.Close()
	}()

	buf, _ := io.ReadAll(fi)
	_ = yaml.Unmarshal(buf, cfg)

	return cfg
}

func TestInitParallelizer(t *testing.T) {
	cfg := testInitConfig()

	_, err := initParallelizer(context.Background(), cfg)
	assert.Equal(t, nil, err)
}

func TestInitPlugin(t *testing.T) {
	cfg := testInitConfig()

	_, err := initPlugin(context.Background(), cfg)
	assert.Equal(t, nil, err)
}

func TestInitScheduler(t *testing.T) {
	cfg := testInitConfig()

	_, err := initScheduler(context.Background(), cfg, nil, nil)
	assert.Equal(t, nil, err)
}

func TestInitServer(t *testing.T) {
	cfg := testInitConfig()

	_, err := initServer(context.Background(), cfg, nil)
	assert.Equal(t, nil, err)
}
