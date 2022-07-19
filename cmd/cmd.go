package cmd

import (
	"context"
	"log"
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
		cobra.CheckErr(run(context.Background()))
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

func run(ctx context.Context) error {
	cfg := config.New()

	if err := initConfig(ctx, cfg); err != nil {
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

func initConfig(_ context.Context, cfg *config.Config) error {
	helper := func(cfg *config.Config) {
		_ = viper.ReadInConfig()
		_ = viper.Unmarshal(cfg)
	}

	helper(cfg)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		helper(cfg)
	})

	return nil
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
			log.Fatalf("failed to run: %v", err)
		}
	}()

	s := make(chan os.Signal, 1)

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can"t be caught, so don't need add it
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		<-s
		_ = srv.Deinit(ctx)
		done <- true
	}()

	<-done

	return nil
}
