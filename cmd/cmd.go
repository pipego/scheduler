package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/parallelizer"
	"github.com/pipego/scheduler/plugin"
	"github.com/pipego/scheduler/scheduler"
	"github.com/pipego/scheduler/server"
)

var (
	configFile string
	listenUrl  string
)

var rootCmd = &cobra.Command{
	Use:     "scheduler",
	Version: config.Version + "-build-" + config.Build,
	Short:   "pipego scheduler",
	Long:    `pipego scheduler`,
	Run: func(cmd *cobra.Command, args []string) {
		if configFile == "" || listenUrl == "" {
			_ = cmd.Help()
			return
		}
		cobra.CheckErr(loadConfig())
	},
}

// nolint: gochecknoinits
func init() {
	helper := func() {
		if configFile != "" {
			viper.SetConfigFile(configFile)
		} else {
			home, _ := homedir.Dir()
			viper.AddConfigPath(home)
			viper.AddConfigPath(".")
			viper.SetConfigName("config")
			viper.SetConfigType("yml")
		}
	}

	cobra.OnInitialize(helper)

	rootCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "config file (.yml)")
	_ = rootCmd.MarkFlagRequired("config-file")

	rootCmd.Flags().StringVarP(&listenUrl, "listen-url", "l", "", "listen url (host:port)")
	_ = rootCmd.MarkFlagRequired("listen-url")
}

func Execute() error {
	return rootCmd.Execute()
}

func loadConfig() error {
	helper := func(ctx context.Context, cfg *config.Config) (server.Server, error) {
		if err := viper.ReadInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to read config")
		}
		if err := viper.Unmarshal(cfg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal config")
		}
		srv, err := initPipe(ctx, cfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init pipe")
		}
		return srv, nil
	}

	cfg := config.New()
	ctx := context.Background()
	reload := make(chan bool, 1)

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can"t be caught, so don't need add it
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	srv, err := helper(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to load")
	}

	if err = runPipe(ctx, srv); err != nil {
		return errors.Wrap(err, "failed to run")
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		reload <- true
	})

L:
	for {
		select {
		case <-reload:
			_ = stopPipe(ctx, srv)
			srv, err = helper(ctx, cfg)
			if err != nil {
				return errors.Wrap(err, "failed to reload")
			}
			if err := runPipe(ctx, srv); err != nil {
				return errors.Wrap(err, "failed to run")
			}
		case <-sig:
			_ = stopPipe(ctx, srv)
			break L
		}
	}

	return nil
}

func initPipe(ctx context.Context, cfg *config.Config) (server.Server, error) {
	pa, err := initParallelizer(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init parallelizer")
	}

	pl, err := initPlugin(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init plugin")
	}

	sched, err := initScheduler(ctx, cfg, pa, pl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init scheduler")
	}

	srv, err := initServer(ctx, cfg, sched)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init server")
	}

	return srv, nil
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

	c.Address = listenUrl
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
			fmt.Println("failed to run")
			return
		}
	}()

	return nil
}

func stopPipe(ctx context.Context, srv server.Server) error {
	return srv.Deinit(ctx)
}
