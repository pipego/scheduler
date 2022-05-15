package cmd

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v3"

	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/plugin"
	"github.com/pipego/scheduler/scheduler"
	"github.com/pipego/scheduler/server"
)

var (
	app        = kingpin.New("scheduler", "pipego scheduler").Version(config.Version + "-build-" + config.Build)
	configFile = app.Flag("config-file", "Config file (.yml)").Required().String()
	listenUrl  = app.Flag("listen-url", "Listen URL (host:port)").Required().String()
)

func Run() error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	cfg, err := initConfig(*configFile)
	if err != nil {
		return errors.Wrap(err, "failed to init config")
	}

	pl, err := initPlugin(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init plugin")
	}

	sched, err := initScheduler(cfg, pl)
	if err != nil {
		return errors.Wrap(err, "failed to init scheduler")
	}

	srv, err := initServer(cfg, sched)
	if err != nil {
		return errors.Wrap(err, "failed to init server")
	}

	log.Println("running")

	if err := runPipe(srv); err != nil {
		return errors.Wrap(err, "failed to run pipe")
	}

	log.Println("exiting")

	return nil
}

func initConfig(name string) (*config.Config, error) {
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

func initPlugin(cfg *config.Config) (plugin.Plugin, error) {
	c := plugin.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg

	return plugin.New(context.Background(), c), nil
}

func initScheduler(cfg *config.Config, pl plugin.Plugin) (scheduler.Scheduler, error) {
	c := scheduler.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Plugin = pl

	return scheduler.New(context.Background(), c), nil
}

func initServer(cfg *config.Config, sched scheduler.Scheduler) (server.Server, error) {
	c := server.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Address = *listenUrl
	c.Config = *cfg
	c.Scheduler = sched

	return server.New(context.Background(), c), nil
}

func runPipe(srv server.Server) error {
	if err := srv.Init(); err != nil {
		return errors.Wrap(err, "failed to init")
	}

	if err := srv.Run(); err != nil {
		return errors.Wrap(err, "failed to run")
	}

	return nil
}
