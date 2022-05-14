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
	if err != nil {
		return errors.Wrap(err, "failed to init filter")
	}

	p.filter = map[string]FilterImpl{}
	for k, v := range buf {
		p.filter[k] = v.(FilterImpl)
	}

	buf, err = initPlugin(&p.cfg.Config.Spec.Score, &Score{})
	if err != nil {
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

	return p.fetch[name].Run(host)
}

func (p *plugin) RunFilter(name string, args *common.Args) FilterResult {
	if _, ok := p.filter[name]; !ok {
		return FilterResult{Error: "invalid name"}
	}

	return p.filter[name].Run(args)
}

func (p *plugin) RunScore(name string, args *common.Args) ScoreResult {
	if _, ok := p.score[name]; !ok {
		return ScoreResult{Score: -1}
	}

	return p.score[name].Run(args)
}
