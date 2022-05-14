package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
)

func TestInitPlugin(t *testing.T) {
	cfg := config.Plugin{}

	buf, err := initPlugin(&cfg, &Fetch{})
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(buf))

	cfg.Disabled = []config.Disabled{
		{
			Name: "name1",
		},
	}

	cfg.Enabled = []config.Enabled{}

	_, err = initPlugin(&cfg, &Fetch{})
	assert.NotEqual(t, nil, err)

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "plugin/fetch-localhost",
		},
		{
			Name: "LocalHost",
			Path: "plugin/fetch-localhost",
		},
	}

	_, err = initPlugin(&cfg, &Fetch{})
	assert.NotEqual(t, nil, err)

	cfg.Disabled = []config.Disabled{
		{
			Name: "LocalHost",
			Path: "plugin/fetch-localhost",
		},
	}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "plugin/fetch-localhost",
		},
	}

	_, err = initPlugin(&cfg, &Fetch{})
	assert.NotEqual(t, nil, err)

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "plugin/fetch-localhost",
		},
	}

	buf, err = initPlugin(&cfg, &Fetch{})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, len(buf))
}

func TestRunFetch(t *testing.T) {
	cfg := config.Plugin{}

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "plugin/fetch-localhost",
		},
	}

	buf, _ := initPlugin(&cfg, &Fetch{})

	pl := plugin{}

	for k, v := range buf {
		pl.fetch[k] = v.(FetchRoutine)
	}

	res := pl.RunFetch("invalid", "")
	assert.LessOrEqual(t, 0, res.AllocatableResource.MilliCPU)

	res = pl.RunFetch("LocalHost", "127.0.0.1")
	assert.Greater(t, 0, res.AllocatableResource.MilliCPU)
}

func TestRunFilter(t *testing.T) {
	cfg := config.Plugin{}

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "NodeName",
			Path: "plugin/filter-nodename",
		},
	}

	buf, _ := initPlugin(&cfg, &Filter{})

	pl := plugin{}

	for k, v := range buf {
		pl.filter[k] = v.(FilterRoutine)
	}

	res := pl.RunFilter("invalid", nil)
	assert.NotEqual(t, "", res.Error)

	args := &common.Args{}

	res = pl.RunFilter("NodeName", args)
	assert.Equal(t, "", res.Error)
}

func TestRunScore(t *testing.T) {
	cfg := config.Plugin{}

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "NodeResourcesFit",
			Path: "plugin/score-noderesourcesfit",
		},
	}

	buf, _ := initPlugin(&cfg, &Score{})

	pl := plugin{}

	for k, v := range buf {
		pl.score[k] = v.(ScoreRoutine)
	}

	res := pl.RunScore("invalid", nil)
	assert.Less(t, 0, res.Score)

	args := &common.Args{}

	res = pl.RunScore("NodeResourcesFit", args)
	assert.GreaterOrEqual(t, 0, res.Score)
}
