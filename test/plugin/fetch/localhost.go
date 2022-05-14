package main

import (
	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/plugin"
)

type LocalHost struct{}

func (t *LocalHost) Run(_ string) plugin.FetchResult {
	return plugin.FetchResult{
		AllocatableResource: common.Resource{
			MilliCPU: 100,
		},
	}
}

// nolint:typecheck
func main() {
	config := gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	pluginMap := map[string]gop.Plugin{
		"LocalHost": &plugin.Fetch{Impl: &LocalHost{}},
	}

	gop.Serve(&gop.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
