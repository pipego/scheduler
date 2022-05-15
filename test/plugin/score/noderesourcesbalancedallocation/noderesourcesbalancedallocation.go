package main

import (
	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/plugin"
)

type NodeResourcesBalancedAllocation struct{}

func (t *NodeResourcesBalancedAllocation) Run(_ *common.Args) plugin.ScoreResult {
	return plugin.ScoreResult{}
}

// nolint:typecheck
func main() {
	config := gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	pluginMap := map[string]gop.Plugin{
		"NodeResourcesBalancedAllocation": &plugin.Score{Impl: &NodeResourcesBalancedAllocation{}},
	}

	gop.Serve(&gop.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
