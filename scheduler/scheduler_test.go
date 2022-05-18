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

	s := scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init()
	_, err := s.runFetchPlugins(nodes)
	assert.Equal(t, nil, err)
	_ = s.Deinit()

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

	_ = s.Init()
	_, err = s.runFetchPlugins(nodes)
	assert.NotEqual(t, nil, err)
	_ = s.Deinit()

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

	_ = s.Init()
	nodes = append(nodes, &common.Node{Host: "127.0.0.1"})
	buf, err := s.runFetchPlugins(nodes)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(100), buf[0].AllocatableResource.MilliCPU)
	_ = s.Deinit()
}

func TestRunFilterPlugins(t *testing.T) {
	var task common.Task
	var nodes []*common.Node

	s := scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init()
	_, err := s.runFilterPlugins(&task, nodes)
	assert.Equal(t, nil, err)
	_ = s.Deinit()

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

	_ = s.Init()
	nodes = append(nodes, &common.Node{Host: "127.0.0.1"})
	buf, err := s.runFilterPlugins(&task, nodes)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(buf))
	_ = s.Deinit()
}

func TestRunScorePlugins(t *testing.T) {
	var task common.Task
	var nodes []*common.Node

	s := scheduler{
		cfg: &Config{
			Config: cfg,
			Plugin: initPlugin(&cfg),
		},
	}

	_ = s.Init()
	_, err := s.runScorePlugins(&task, nodes)
	assert.NotEqual(t, nil, err)
	_ = s.Deinit()

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

	_ = s.Init()
	nodes = append(nodes, &common.Node{Host: "127.0.0.1"})
	buf, err := s.runScorePlugins(&task, nodes)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(buf))
	_ = s.Deinit()
}

func TestSelectHost(t *testing.T) {
	var scores []nodeScore

	s := scheduler{}

	_, err := s.selectHost(scores)
	assert.NotEqual(t, nil, err)

	scores = append(scores, nodeScore{
		name:  "name1",
		score: 0,
	})

	buf, err := s.selectHost(scores)
	assert.Equal(t, nil, err)
	assert.Equal(t, "name1", buf)

	scores = append(scores, nodeScore{
		name:  "name2",
		score: 1,
	})

	buf, err = s.selectHost(scores)
	assert.Equal(t, nil, err)
	assert.Equal(t, "name2", buf)

	scores = append(scores, nodeScore{
		name:  "name3",
		score: 1,
	})

	buf, err = s.selectHost(scores)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", buf)
}
