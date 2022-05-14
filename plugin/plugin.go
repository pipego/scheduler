package plugin

import (
	"context"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	gop "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
)

type Plugin interface {
	Init() error
	RunFetch(string, string) FetchResult
	RunFilter(string, *common.Args) FilterResult
	RunScore(string, *common.Args) ScoreResult
}

type FetchImpl interface {
	Run(string) FetchResult
}

type FilterImpl interface {
	Run(args *common.Args) FilterResult
}

type ScoreImpl interface {
	Run(args *common.Args) ScoreResult
}

type Config struct {
	Config config.Config
}

type FetchRoutine func(string) FetchResult

type FilterRoutine func(args *common.Args) FilterResult

type ScoreRoutine func(args *common.Args) ScoreResult

type FetchResult struct {
	AllocatableResource common.Resource
	RequestedResource   common.Resource
}

type FilterResult struct {
	Error string
}

type ScoreResult struct {
	Score int64
}

type plugin struct {
	cfg    *Config
	fetch  map[string]FetchRoutine
	filter map[string]FilterRoutine
	score  map[string]ScoreRoutine
}

var (
	handshake = gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stderr,
		Level:  hclog.Error,
	})
)

func New(_ context.Context, cfg *Config) Plugin {
	return &plugin{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *plugin) Init() error {
	var err error

	buf, err := initPlugin(&p.cfg.Config.Spec.Fetch, &Fetch{})
	if err != nil {
		return errors.Wrap(err, "failed to init fetch")
	}

	for k, v := range buf {
		p.fetch[k] = v.(FetchRoutine)
	}

	buf, err = initPlugin(&p.cfg.Config.Spec.Filter, &Filter{})
	if err != nil {
		return errors.Wrap(err, "failed to init filter")
	}

	for k, v := range buf {
		p.filter[k] = v.(FilterRoutine)
	}

	buf, err = initPlugin(&p.cfg.Config.Spec.Score, &Score{})
	if err != nil {
		return errors.Wrap(err, "failed to init score")
	}

	for k, v := range buf {
		p.score[k] = v.(ScoreRoutine)
	}

	return nil
}

func initPlugin(cfg *config.Plugin, impl gop.Plugin) (map[string]interface{}, error) {
	var err error
	pl := make(map[string]interface{})

	for _, item := range cfg.Disabled {
		if _, ok := pl[item.Name]; ok {
			return nil, errors.New("duplicate name")
		}
		pl[item.Name], err = initHelper(item.Name, item.Path, impl)
		if err != nil {
			return nil, errors.Wrap(err, "failed to register disabled")
		}
	}

	for _, item := range cfg.Enabled {
		if _, ok := pl[item.Name]; ok {
			return nil, errors.New("duplicate name")
		}
		pl[item.Name], err = initHelper(item.Name, item.Path, impl)
		if err != nil {
			return nil, errors.Wrap(err, "failed to register enabled")
		}
	}

	return pl, nil
}

func initHelper(name, path string, impl gop.Plugin) (interface{}, error) {
	plugins := map[string]gop.Plugin{
		name: impl,
	}

	client := gop.NewClient(&gop.ClientConfig{
		Cmd:             exec.Command(path),
		HandshakeConfig: handshake,
		Logger:          logger,
		Plugins:         plugins,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create")
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dispense")
	}

	return raw, nil
}

func (p *plugin) RunFetch(name, host string) FetchResult {
	if _, ok := p.fetch[name]; !ok {
		return FetchResult{}
	}

	return p.fetch[name](host)
}

func (p *plugin) RunFilter(name string, args *common.Args) FilterResult {
	if _, ok := p.filter[name]; !ok {
		return FilterResult{Error: "invalid name"}
	}

	return p.filter[name](args)
}

func (p *plugin) RunScore(name string, args *common.Args) ScoreResult {
	if _, ok := p.score[name]; !ok {
		return ScoreResult{Score: -1}
	}

	return p.score[name](args)
}
