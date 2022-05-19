package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/plugin"
)

var (
	cfg = config.Config{
		Spec: config.Spec{
			Fetch: config.Plugin{
				Enabled: []config.Enabled{},
			},
			Filter: config.Plugin{
				Enabled: []config.Enabled{},
			},
			Score: config.Plugin{
				Enabled: []config.Enabled{},
			},
		},
	}
)

func initPlugin(cfg *config.Config) plugin.Plugin {
	c := plugin.DefaultConfig()
	c.Config = *cfg

	return plugin.New(context.Background(), c)
}

func TestRunFetchPlugins(t *testing.T) {
	var nodes []*common.Node
	ctx := context.Background()

	s := scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	_, err := s.runFetchPlugins(ctx, nodes)
	assert.Equal(t, nil, err)
	_ = s.Deinit(ctx)

	cfg.Spec.Fetch.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "../plugin/fetch-localhost",
		},
		{
			Name: "LocalHost",
			Path: "../plugin/fetch-localhost",
		},
	}

	s = scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	_, err = s.runFetchPlugins(ctx, nodes)
	assert.NotEqual(t, nil, err)
	_ = s.Deinit(ctx)

	cfg.Spec.Fetch.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "../plugin/fetch-localhost",
		},
	}

	s = scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	nodes = append(nodes, &common.Node{Host: "127.0.0.1"})
	buf, err := s.runFetchPlugins(ctx, nodes)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(100), buf[0].AllocatableResource.MilliCPU)
	_ = s.Deinit(ctx)
}

func TestRunFilterPlugins(t *testing.T) {
	var task common.Task
	var nodes []*common.Node
	ctx := context.Background()

	s := scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	_, err := s.runFilterPlugins(ctx, &task, nodes)
	assert.Equal(t, nil, err)
	_ = s.Deinit(ctx)

	cfg.Spec.Filter.Enabled = []config.Enabled{
		{
			Name:     "NodeAffinity",
			Path:     "../plugin/filter-nodeaffinity",
			Priority: 2,
		},
		{
			Name:     "NodeName",
			Path:     "../plugin/filter-nodename",
			Priority: 1,
		},
	}

	s = scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	nodes = append(nodes, &common.Node{Host: "127.0.0.1"})
	buf, err := s.runFilterPlugins(ctx, &task, nodes)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(buf))
	_ = s.Deinit(ctx)
}

func TestRunScorePlugins(t *testing.T) {
	var task common.Task
	var nodes []*common.Node
	ctx := context.Background()

	s := scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	_, err := s.runScorePlugins(ctx, &task, nodes)
	assert.NotEqual(t, nil, err)
	_ = s.Deinit(ctx)

	cfg.Spec.Score.Enabled = []config.Enabled{
		{
			Name:   "NodeResourcesFit",
			Path:   "../plugin/score-noderesourcesfit",
			Weight: 2,
		},
		{
			Name:   "NodeResourcesBalancedAllocation",
			Path:   "../plugin/score-noderesourcesbalancedallocation",
			Weight: 1,
		},
	}

	s = scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init(ctx)
	nodes = append(nodes, &common.Node{Host: "127.0.0.1"})
	buf, err := s.runScorePlugins(ctx, &task, nodes)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(buf))
	_ = s.Deinit(ctx)
}

func TestSelectHost(t *testing.T) {
	var scores []nodeScore
	ctx := context.Background()

	s := scheduler{}

	_, err := s.selectHost(ctx, scores)
	assert.NotEqual(t, nil, err)

	scores = append(scores, nodeScore{
		name:  "name1",
		score: 0,
	})

	buf, err := s.selectHost(ctx, scores)
	assert.Equal(t, nil, err)
	assert.Equal(t, "name1", buf)

	scores = append(scores, nodeScore{
		name:  "name2",
		score: 1,
	})

	buf, err = s.selectHost(ctx, scores)
	assert.Equal(t, nil, err)
	assert.Equal(t, "name2", buf)

	scores = append(scores, nodeScore{
		name:  "name3",
		score: 1,
	})

	buf, err = s.selectHost(ctx, scores)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", buf)
}
