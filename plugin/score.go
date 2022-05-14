package plugin

import (
	"net/rpc"

	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/common"
)

type ScoreRPC struct {
	client *rpc.Client
}

func (n *ScoreRPC) Run(args *common.Args) ScoreResult {
	var resp ScoreResult
	if err := n.client.Call("Plugin.Run", args, &resp); err != nil {
		panic(err)
	}
	return resp
}

type ScoreRPCServer struct {
	Impl ScoreImpl
}

func (n *ScoreRPCServer) Run(args *common.Args, resp *ScoreResult) error {
	*resp = n.Impl.Run(args)
	return nil
}

type Score struct {
	Impl ScoreImpl
}

func (n *Score) Server(*gop.MuxBroker) (interface{}, error) {
	return &ScoreRPCServer{Impl: n.Impl}, nil
}

func (Score) Client(b *gop.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ScoreRPC{client: c}, nil
}
