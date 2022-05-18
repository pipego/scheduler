package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
)

func TestInitPlugin(t *testing.T) {
	cfg := config.Plugin{}
	pl := plugin{}

	c, p, err := pl.initPlugin(&cfg, &Fetch{})
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(c))
	assert.Equal(t, 0, len(p))
	_ = pl.deinitHelper(c)

	cfg.Disabled = []config.Disabled{
		{
			Name: "name1",
		},
	}

	cfg.Enabled = []config.Enabled{}

	c, p, err = pl.initPlugin(&cfg, &Fetch{})
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 0, len(c))
	assert.Equal(t, 0, len(p))
	_ = pl.deinitHelper(c)

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "./fetch-localhost",
		},
		{
			Name: "LocalHost",
			Path: "./fetch-localhost",
		},
	}

	c, p, err = pl.initPlugin(&cfg, &Fetch{})
	assert.NotEqual(t, nil, err)
	assert.NotEqual(t, 0, len(c))
	assert.NotEqual(t, 0, len(p))
	_ = pl.deinitHelper(c)

	cfg.Disabled = []config.Disabled{
		{
			Name: "LocalHost",
			Path: "./fetch-localhost",
		},
	}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "./fetch-localhost",
		},
	}

	c, p, err = pl.initPlugin(&cfg, &Fetch{})
	assert.NotEqual(t, nil, err)
	assert.NotEqual(t, 0, len(c))
	assert.NotEqual(t, 0, len(p))
	_ = pl.deinitHelper(c)

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "./fetch-localhost",
		},
	}

	c, p, err = pl.initPlugin(&cfg, &Fetch{})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, len(c))
	assert.NotEqual(t, 0, len(p))
	_ = pl.deinitHelper(c)
}

func TestInitInstance(t *testing.T) {
	pl := plugin{}

	name := ""
	_path := ""

	_, _, err := pl.initInstance(name, _path, &Fetch{})
	assert.NotEqual(t, nil, err)

	name = "name1"
	_path = ""

	_, _, err = pl.initInstance(name, _path, &Fetch{})
	assert.NotEqual(t, nil, err)

	name = ""
	_path = "path1"

	_, _, err = pl.initInstance(name, _path, &Fetch{})
	assert.NotEqual(t, nil, err)

	name = "LocalHost"
	_path = "./fetch-localhost"

	c, _, err := pl.initInstance(name, _path, &Fetch{})
	assert.Equal(t, nil, err)
	c.Kill()
}

func TestRunFetch(t *testing.T) {
	cfg := config.Plugin{}
	pl := plugin{}

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "LocalHost",
			Path: "./fetch-localhost",
		},
	}

	c, p, _ := pl.initPlugin(&cfg, &Fetch{})

	pl.fetch = map[string]FetchImpl{}
	for k, v := range p {
		pl.fetch[k] = v.(*FetchRPC)
	}

	_, err := pl.RunFetch("invalid", "")
	assert.NotEqual(t, nil, err)

	_, err = pl.RunFetch("LocalHost", "127.0.0.1")
	assert.Equal(t, nil, err)

	_ = pl.deinitHelper(c)
}

func TestRunFilter(t *testing.T) {
	cfg := config.Plugin{}
	pl := plugin{}

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "NodeName",
			Path: "./filter-nodename",
		},
	}

	c, p, _ := pl.initPlugin(&cfg, &Filter{})

	pl.filter = map[string]FilterImpl{}
	for k, v := range p {
		pl.filter[k] = v.(*FilterRPC)
	}

	task := common.Task{}
	node := common.Node{}

	_, err := pl.RunFilter("invalid", &task, &node)
	assert.NotEqual(t, nil, err)

	_, err = pl.RunFilter("NodeName", &task, &node)
	assert.Equal(t, nil, err)

	_ = pl.deinitHelper(c)
}

func TestRunScore(t *testing.T) {
	cfg := config.Plugin{}
	pl := plugin{}

	cfg.Disabled = []config.Disabled{}

	cfg.Enabled = []config.Enabled{
		{
			Name: "NodeResourcesFit",
			Path: "./score-noderesourcesfit",
		},
	}

	c, p, _ := pl.initPlugin(&cfg, &Score{})

	pl.score = map[string]ScoreImpl{}
	for k, v := range p {
		pl.score[k] = v.(*ScoreRPC)
	}

	task := common.Task{}
	node := common.Node{}

	_, err := pl.RunScore("invalid", &task, &node)
	assert.NotEqual(t, nil, err)

	_, err = pl.RunScore("NodeResourcesFit", &task, &node)
	assert.Equal(t, nil, err)

	_ = pl.deinitHelper(c)
}
