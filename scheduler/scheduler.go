package scheduler

import (
	"context"
	"math/rand"
	"sort"

	"github.com/pkg/errors"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/plugin"
)

type Scheduler interface {
	Init() error
	Run(*common.Task, []*common.Node) Result
}

type Config struct {
	Config config.Config
	Plugin plugin.Plugin
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

func (s *scheduler) Init() error {
	if err := s.cfg.Plugin.Init(); err != nil {
		return errors.Wrap(err, "failed to init plugin")
	}

	return nil
}

func (s *scheduler) Run(task *common.Task, nodes []*common.Node) Result {
	var scores []nodeScore

	if len(nodes) == 0 {
		return Result{Error: "invalid nodes"}
	}

	nodes, err := s.runFetchPlugins(nodes)
	if err != nil {
		return Result{Error: "failed to fetch"}
	}

	nodes, err = s.runFilterPlugins(task, nodes)
	if err != nil {
		return Result{Error: "failed to filter"}
	}

	scores, err = s.runScorePlugins(task, nodes)
	if err != nil {
		return Result{Error: "failed to score"}
	}

	host, err := s.selectHost(scores)
	if err != nil {
		return Result{Error: "failed to select"}
	}

	return Result{Name: host}
}

func (s *scheduler) runFetchPlugins(nodes []*common.Node) ([]*common.Node, error) {
	helper := func(node *common.Node, res plugin.FetchResult) *common.Node {
		if res.AllocatableResource.MilliCPU < 0 ||
			res.AllocatableResource.Memory < 0 ||
			res.AllocatableResource.Storage < 0 ||
			res.RequestedResource.MilliCPU < 0 ||
			res.RequestedResource.Memory < 0 ||
			res.RequestedResource.Storage < 0 {
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

	// TODO: Set in parallel
	for i := range nodes {
		if res, err := s.cfg.Plugin.RunFetch(pl.Name, nodes[i].Host); err == nil {
			nodes[i] = helper(nodes[i], res)
		}
	}

	return nodes, nil
}

func (s *scheduler) runFilterPlugins(task *common.Task, nodes []*common.Node) ([]*common.Node, error) {
	var buf []*common.Node

	filterHelper := func(p string, t *common.Task, n []*common.Node) []*common.Node {
		var b []*common.Node
		for i := range n {
			if res, err := s.cfg.Plugin.RunFilter(p, t, n[i]); err == nil {
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
		return pl[i].Weight > pl[j].Weight
	})

	// TODO: Set in parallel
	for _, item := range pl {
		buf = filterHelper(item.Name, task, nodes)
		if len(buf) != 0 {
			break
		}
	}

	return buf, nil
}

func (s *scheduler) runScorePlugins(task *common.Task, nodes []*common.Node) ([]nodeScore, error) {
	// TODO: Add implementation
	return nil, nil
}

// nolint: gosec
func (s *scheduler) selectHost(scores []nodeScore) (string, error) {
	if len(scores) == 0 {
		return "", errors.New("invalid scores")
	}

	count := 1
	selected := scores[0].name
	max := scores[0].score

	for _, item := range scores[1:] {
		if item.score > max {
			max = item.score
			selected = item.name
			count = 1
		} else if item.score == max {
			count++
			if rand.Intn(count) == 0 {
				// Replace the candidate with probability of 1/count
				selected = item.name
			}
		}
	}

	return selected, nil
}
