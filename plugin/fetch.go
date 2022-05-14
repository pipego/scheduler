package plugin

import (
	"net/rpc"

	gop "github.com/hashicorp/go-plugin"
)

type FetchRPC struct {
	client *rpc.Client
}

func (n *FetchRPC) Run(host string) FetchResult {
	var resp FetchResult
	if err := n.client.Call("Plugin.Run", host, &resp); err != nil {
		panic(err)
	}
	return resp
}

type FetchRPCServer struct {
	Impl FetchImpl
}

func (n *FetchRPCServer) Run(host string, resp *FetchResult) error {
	*resp = n.Impl.Run(host)
	return nil
}

type Fetch struct {
	Impl FetchImpl
}

func (n *Fetch) Server(*gop.MuxBroker) (interface{}, error) {
	return &FetchRPCServer{Impl: n.Impl}, nil
}

func (Fetch) Client(b *gop.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FetchRPC{client: c}, nil
}
