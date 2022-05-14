package plugin

import (
	"net/rpc"

	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/common"
)

type FilterRPC struct {
	client *rpc.Client
}

func (n *FilterRPC) Run(args *common.Args) FilterResult {
	var resp FilterResult
	if err := n.client.Call("Plugin.Run", args, &resp); err != nil {
		panic(err)
	}
	return resp
}

type FilterRPCServer struct {
	Impl FilterImpl
}

func (n *FilterRPCServer) Run(args *common.Args, resp *FilterResult) error {
	*resp = n.Impl.Run(args)
	return nil
}

type Filter struct {
	Impl FilterImpl
}

func (n *Filter) Server(*gop.MuxBroker) (interface{}, error) {
	return &FilterRPCServer{Impl: n.Impl}, nil
}

func (Filter) Client(b *gop.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FilterRPC{client: c}, nil
}
