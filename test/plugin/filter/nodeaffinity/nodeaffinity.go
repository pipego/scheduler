package main

import (
	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/plugin"
)

type NodeAffinity struct{}

func (t *NodeAffinity) Run(_ *common.Args) plugin.FilterResult {
	return plugin.FilterResult{}
}

// nolint:typecheck
func main() {
	config := gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	pluginMap := map[string]gop.Plugin{
		"NodeAffinity": &plugin.Filter{Impl: &NodeAffinity{}},
	}

	gop.Serve(&gop.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
