package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v3"

	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/parallelizer"
	"github.com/pipego/scheduler/plugin"
	"github.com/pipego/scheduler/scheduler"
	"github.com/pipego/scheduler/server"
)

const (
	timeout = 5 * time.Second
)

var (
	app        = kingpin.New("scheduler", "pipego scheduler").Version(config.Version + "-build-" + config.Build)
	configFile = app.Flag("config-file", "Config file (.yml)").Required().String()
	listenUrl  = app.Flag("listen-url", "Listen URL (host:port)").Required().String()
)

func Run(ctx context.Context) error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	cfg, err := initConfig(ctx, *configFile)
	if err != nil {
		return errors.Wrap(err, "failed to init config")
	}

	pa, err := initParallelizer(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init parallelizer")
	}

	pl, err := initPlugin(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init plugin")
	}

	sched, err := initScheduler(ctx, cfg, pa, pl)
	if err != nil {
		return errors.Wrap(err, "failed to init scheduler")
	}

	srv, err := initServer(ctx, cfg, sched)
	if err != nil {
		return errors.Wrap(err, "failed to init server")
	}

	if err := runPipe(ctx, srv); err != nil {
		return errors.Wrap(err, "failed to run pipe")
	}

	return nil
}

func initConfig(_ context.Context, name string) (*config.Config, error) {
	c := config.New()

	fi, err := os.Open(name)
	if err != nil {
		return c, errors.Wrap(err, "failed to open")
	}

	defer func() {
		_ = fi.Close()
	}()

	buf, _ := io.ReadAll(fi)

	if err := yaml.Unmarshal(buf, c); err != nil {
		return c, errors.Wrap(err, "failed to unmarshal")
	}

	return c, nil
}

func initParallelizer(ctx context.Context, cfg *config.Config) (parallelizer.Parallelizer, error) {
	c := parallelizer.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg

	return parallelizer.New(ctx, c), nil
}

func initPlugin(ctx context.Context, cfg *config.Config) (plugin.Plugin, error) {
	c := plugin.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg

	return plugin.New(ctx, c), nil
}

func initScheduler(ctx context.Context, cfg *config.Config, pa parallelizer.Parallelizer, pl plugin.Plugin) (scheduler.Scheduler, error) {
	c := scheduler.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Parallelizer = pa
	c.Plugin = pl

	return scheduler.New(ctx, c), nil
}

func initServer(ctx context.Context, cfg *config.Config, sched scheduler.Scheduler) (server.Server, error) {
	c := server.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Address = *listenUrl
	c.Config = *cfg
	c.Scheduler = sched

	return server.New(ctx, c), nil
}

func runPipe(ctx context.Context, srv server.Server) error {
	if err := srv.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init")
	}

	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Fatalf("failed to run: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can"t be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_ = srv.Deinit(c)
	<-c.Done()

	return nil
}
