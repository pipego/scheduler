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
	Deinit() error
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
	client []*gop.Client
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

	cli, pl, err := p.initPlugin(&p.cfg.Config.Spec.Fetch, &Fetch{})
	if err != nil {
		_ = p.deinitHelper(cli)
		return errors.Wrap(err, "failed to init fetch")
	}

	p.client = append(p.client, cli...)

	p.fetch = map[string]FetchImpl{}
	for k, v := range pl {
		p.fetch[k] = v.(FetchImpl)
	}

	cli, pl, err = p.initPlugin(&p.cfg.Config.Spec.Filter, &Filter{})
	if err != nil || len(pl) == 0 {
		_ = p.deinitHelper(cli)
		return errors.Wrap(err, "failed to init filter")
	}

	p.client = append(p.client, cli...)

	p.filter = map[string]FilterImpl{}
	for k, v := range pl {
		p.filter[k] = v.(FilterImpl)
	}

	cli, pl, err = p.initPlugin(&p.cfg.Config.Spec.Score, &Score{})
	if err != nil || len(pl) == 0 {
		_ = p.deinitHelper(cli)
		return errors.Wrap(err, "failed to init score")
	}

	p.client = append(p.client, cli...)

	p.score = map[string]ScoreImpl{}
	for k, v := range pl {
		p.score[k] = v.(ScoreImpl)
	}

	return nil
}

func (p *plugin) initPlugin(cfg *config.Plugin, impl gop.Plugin) ([]*gop.Client, map[string]interface{}, error) {
	var cli []*gop.Client
	pl := make(map[string]interface{})

	helper := func(name, _path string) error {
		if _, ok := pl[name]; ok {
			return errors.New("duplicate name")
		}
		c, i, err := p.initInstance(name, _path, impl)
		if err != nil {
			return errors.New("failed to init instance")
		}
		cli = append(cli, c)
		pl[name] = i
		return nil
	}

	for _, item := range cfg.Disabled {
		if err := helper(item.Name, item.Path); err != nil {
			return cli, pl, err
		}
	}

	for _, item := range cfg.Enabled {
		if err := helper(item.Name, item.Path); err != nil {
			return cli, pl, err
		}
	}

	return cli, pl, nil
}

func (p *plugin) initInstance(name, _path string, impl gop.Plugin) (*gop.Client, interface{}, error) {
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
		return nil, nil, errors.Wrap(err, "failed to create")
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		client.Kill()
		return nil, nil, errors.Wrap(err, "failed to dispense")
	}

	return client, raw, nil
}

func (p *plugin) Deinit() error {
	return p.deinitHelper(p.client)
}

func (p *plugin) deinitHelper(cli []*gop.Client) error {
	for _, item := range cli {
		item.Kill()
	}

	return nil
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
