package server

import (
	"context"
	"math"
	"net"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	pb "github.com/pipego/scheduler/server/proto"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/scheduler"
)

const (
	KIND = "scheduler"
)

type Server interface {
	Init() error
	Run() error
}

type Config struct {
	Address   string
	Config    config.Config
	Scheduler scheduler.Scheduler
}

type server struct {
	cfg   *Config
	nodes []*common.Node
	task  *common.Task
}

type rpcServer struct {
	pb.UnimplementedServerProtoServer
}

func New(_ context.Context, cfg *Config) Server {
	return &server{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (s *server) Init() error {
	if err := s.cfg.Scheduler.Init(); err != nil {
		return errors.Wrap(err, "failed to init scheduler")
	}

	return nil
}

func (s *server) Run() error {
	options := []grpc.ServerOption{grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32)}

	g := grpc.NewServer(options...)
	pb.RegisterServerProtoServer(g, &rpcServer{})

	lis, _ := net.Listen("tcp", s.cfg.Address)

	return g.Serve(lis)
}

func (s *server) SendServer(in *pb.ServerRequest) (*pb.ServerReply, error) {
	if in.GetKind() != KIND {
		return &pb.ServerReply{Error: "invalid kind"}, nil
	}

	if err := s.sendHelper(in.GetSpec().GetTask(), in.GetSpec().GetNodes()); err != nil {
		return &pb.ServerReply{Error: "invalid spec"}, nil
	}

	res := s.cfg.Scheduler.Run(s.task, s.nodes)

	return &pb.ServerReply{
		Name:  res.Name,
		Error: res.Error,
	}, nil
}

func (s *server) sendHelper(task *pb.Task, nodes []*pb.Node) error {
	taskHelper := func(t *pb.Task) common.Resource {
		return common.Resource{
			MilliCPU: t.GetRequestedResource().MilliCPU,
			Memory:   t.GetRequestedResource().Memory,
			Storage:  t.GetRequestedResource().Storage,
		}
	}

	nodeHelper := func(n *pb.Node) common.Node {
		return common.Node{
			AllocatableResource: common.Resource{
				MilliCPU: n.GetAllocatableResource().MilliCPU,
				Memory:   n.GetAllocatableResource().Memory,
				Storage:  n.GetAllocatableResource().Storage,
			},
			Host:  n.GetHost(),
			Label: n.GetLabel(),
			Name:  n.GetName(),
			RequestedResource: common.Resource{
				MilliCPU: n.GetRequestedResource().MilliCPU,
				Memory:   n.GetRequestedResource().Memory,
				Storage:  n.GetRequestedResource().Storage,
			},
			Unschedulable: n.GetUnschedulable(),
		}
	}

	s.task = &common.Task{
		Name:                   task.GetName(),
		NodeName:               task.GetNodeName(),
		NodeSelector:           task.GetNodeSelector(),
		RequestedResource:      taskHelper(task),
		ToleratesUnschedulable: task.GetToleratesUnschedulable(),
	}

	for _, item := range nodes {
		node := nodeHelper(item)
		s.nodes = append(s.nodes, &node)
	}

	return nil
}
