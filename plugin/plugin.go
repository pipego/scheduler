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
	RunFetch(string, string) (FetchResult, error)
	RunFilter(string, *common.Task, *common.Node) (FilterResult, error)
	RunScore(string, *common.Task, *common.Node) (ScoreResult, error)
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
	fetch  map[string]FetchImpl
	filter map[string]FilterImpl
	score  map[string]ScoreImpl
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

	p.fetch = map[string]FetchImpl{}
	for k, v := range buf {
		p.fetch[k] = v.(FetchImpl)
	}

	buf, err = initPlugin(&p.cfg.Config.Spec.Filter, &Filter{})
	if err != nil || len(buf) == 0 {
		return errors.Wrap(err, "failed to init filter")
	}

	p.filter = map[string]FilterImpl{}
	for k, v := range buf {
		p.filter[k] = v.(FilterImpl)
	}

	buf, err = initPlugin(&p.cfg.Config.Spec.Score, &Score{})
	if err != nil || len(buf) == 0 {
		return errors.Wrap(err, "failed to init score")
	}

	p.score = map[string]ScoreImpl{}
	for k, v := range buf {
		p.score[k] = v.(ScoreImpl)
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

func initHelper(name, _path string, impl gop.Plugin) (interface{}, error) {
	plugins := map[string]gop.Plugin{
		name: impl,
	}

	client := gop.NewClient(&gop.ClientConfig{
		Cmd:             exec.Command(_path),
		HandshakeConfig: handshake,
		Logger:          logger,
		Plugins:         plugins,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, errors.Wrap(err, "failed to create")
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		client.Kill()
		return nil, errors.Wrap(err, "failed to dispense")
	}

	return raw, nil
}

func (p *plugin) RunFetch(name, host string) (FetchResult, error) {
	if _, ok := p.fetch[name]; !ok {
		return FetchResult{}, errors.New("invalid name")
	}

	return p.fetch[name].Run(host), nil
}

func (p *plugin) RunFilter(name string, task *common.Task, node *common.Node) (FilterResult, error) {
	if _, ok := p.filter[name]; !ok {
		return FilterResult{}, errors.New("invalid name")
	}

	args := &common.Args{
		Node: *node,
		Task: *task,
	}

	return p.filter[name].Run(args), nil
}

func (p *plugin) RunScore(name string, task *common.Task, node *common.Node) (ScoreResult, error) {
	if _, ok := p.score[name]; !ok {
		return ScoreResult{}, errors.New("invalid name")
	}

	args := &common.Args{
		Node: *node,
		Task: *task,
	}

	return p.score[name].Run(args), nil
}
