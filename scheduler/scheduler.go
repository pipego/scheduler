package scheduler

import (
	"context"
	"math/rand"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/parallelizer"
	"github.com/pipego/scheduler/plugin"
)

type Scheduler interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, *common.Task, []*common.Node) Result
}

type Config struct {
	Config       config.Config
	Parallelizer parallelizer.Parallelizer
	Plugin       plugin.Plugin
}

type Result struct {
	Name  string
	Error string
}

type scheduler struct {
	cfg *Config
}

type nodeScore struct {
	name  string
	score int64
}

func New(_ context.Context, cfg *Config) Scheduler {
	return &scheduler{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (s *scheduler) Init(ctx context.Context) error {
	if err := s.cfg.Parallelizer.Init(ctx, parallelizer.DefaultParallelism); err != nil {
		return errors.Wrap(err, "failed to init parallelizer")
	}

	if err := s.cfg.Plugin.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init plugin")
	}

	return nil
}

func (s *scheduler) Deinit(ctx context.Context) error {
	return s.cfg.Plugin.Deinit(ctx)
}

func (s *scheduler) Run(ctx context.Context, task *common.Task, nodes []*common.Node) Result {
	var scores []nodeScore

	if len(nodes) == 0 {
		return Result{Error: "invalid nodes"}
	}

	nodes, err := s.runFetchPlugins(ctx, nodes)
	if err != nil {
		return Result{Error: "failed to fetch"}
	}

	nodes, err = s.runFilterPlugins(ctx, task, nodes)
	if err != nil {
		return Result{Error: "failed to filter"}
	}

	scores, err = s.runScorePlugins(ctx, task, nodes)
	if err != nil {
		return Result{Error: "failed to score"}
	}

	host, err := s.selectHost(ctx, scores)
	if err != nil {
		return Result{Error: "failed to select"}
	}

	return Result{Name: host}
}

func (s *scheduler) runFetchPlugins(ctx context.Context, nodes []*common.Node) ([]*common.Node, error) {
	helper := func(node *common.Node, res plugin.FetchResult) *common.Node {
		if res.AllocatableResource.MilliCPU <= 0 &&
			res.AllocatableResource.Memory <= 0 &&
			res.AllocatableResource.Storage <= 0 &&
			res.RequestedResource.MilliCPU <= 0 &&
			res.RequestedResource.Memory <= 0 &&
			res.RequestedResource.Storage <= 0 {
			return node
		}
		node.AllocatableResource.MilliCPU = res.AllocatableResource.MilliCPU
		node.AllocatableResource.Memory = res.AllocatableResource.Memory
		node.AllocatableResource.Storage = res.AllocatableResource.Storage
		node.RequestedResource.MilliCPU = res.RequestedResource.MilliCPU
		node.RequestedResource.Memory = res.RequestedResource.Memory
		node.RequestedResource.Storage = res.RequestedResource.Storage
		return node
	}

	if len(s.cfg.Config.Spec.Fetch.Enabled) == 0 {
		return nodes, nil
	}

	if len(s.cfg.Config.Spec.Fetch.Enabled) > 1 {
		return nil, errors.New("invalid enabled")
	}

	pl := s.cfg.Config.Spec.Fetch.Enabled[0]

	parallelizer.ParallelizeUntil(ctx, parallelizer.DefaultParallelism, len(nodes), func(index int) {
		if res, err := s.cfg.Plugin.RunFetch(ctx, pl.Name, nodes[index].Host); err == nil {
			nodes[index] = helper(nodes[index], res)
		}
	})

	return nodes, nil
}

func (s *scheduler) runFilterPlugins(ctx context.Context, task *common.Task, nodes []*common.Node) ([]*common.Node, error) {
	var buf []*common.Node

	helper := func(p string, t *common.Task, n []*common.Node) []*common.Node {
		var b []*common.Node
		for i := range n {
			if res, err := s.cfg.Plugin.RunFilter(ctx, p, t, n[i]); err == nil {
				if res.Error == "" {
					b = append(b, n[i])
				}
			}
		}
		return b
	}

	if len(s.cfg.Config.Spec.Filter.Enabled) == 0 {
		return nodes, nil
	}

	pl := s.cfg.Config.Spec.Filter.Enabled
	sort.Slice(pl, func(i, j int) bool {
		return pl[i].Priority < pl[j].Priority
	})

	for _, item := range pl {
		buf = helper(item.Name, task, nodes)
		if len(buf) != 0 {
			break
		}
	}

	return buf, nil
}

func (s *scheduler) runScorePlugins(ctx context.Context, task *common.Task, nodes []*common.Node) ([]nodeScore, error) {
	var buf []nodeScore

	helper := func(c config.Enabled, t *common.Task, n []*common.Node) []nodeScore {
		var b []nodeScore
		for i := range n {
			if res, err := s.cfg.Plugin.RunScore(ctx, c.Name, t, n[i]); err == nil {
				if res.Score >= common.MinNodeScore && res.Score <= common.MaxNodeScore {
					b = append(b, nodeScore{
						name:  n[i].Name,
						score: res.Score * c.Weight,
					})
				}
			}
		}
		return b
	}

	if len(s.cfg.Config.Spec.Score.Enabled) == 0 {
		return nil, errors.New("invalid enabled")
	}

	pl := s.cfg.Config.Spec.Score.Enabled
	m := sync.Mutex{}

	parallelizer.ParallelizeUntil(ctx, parallelizer.DefaultParallelism, len(pl), func(index int) {
		m.Lock()
		defer m.Unlock()
		b := helper(pl[index], task, nodes)
		if len(b) != 0 {
			buf = append(buf, b...)
		}
	})

	return buf, nil
}

// nolint: gosec
func (s *scheduler) selectHost(_ context.Context, scores []nodeScore) (string, error) {
	helper := func(s []nodeScore) map[string]int64 {
		b := make(map[string]int64)
		for _, item := range s {
			if _, ok := b[item.name]; ok {
				b[item.name] += item.score
			} else {
				b[item.name] = item.score
			}
		}
		return b
	}

	if len(scores) == 0 {
		return "", errors.New("invalid scores")
	}

	buf := helper(scores)

	count := 1
	max := int64(-1)
	selected := ""

	for key, val := range buf {
		if val > max {
			max = val
			selected = key
			count = 1
		} else if val == max {
			count++
			if rand.Intn(count) == 0 {
				// Replace the candidate with probability of 1/count
				selected = key
			}
		}
	}

	return selected, nil
}
