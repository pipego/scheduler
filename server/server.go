package server

import (
	"context"
	"math"
	"net"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/logger"
	"github.com/pipego/scheduler/scheduler"
	pb "github.com/pipego/scheduler/server/proto"
)

const (
	Kind = "scheduler"
)

type Server interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Address   string
	Config    config.Config
	Logger    logger.Logger
	Scheduler scheduler.Scheduler
}

type server struct {
	cfg  *Config
	srv  *grpc.Server
	task *common.Task
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

func (s *server) Init(ctx context.Context) error {
	if err := s.cfg.Logger.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init logger")
	}

	if err := s.cfg.Scheduler.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init scheduler")
	}

	options := []grpc.ServerOption{grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32)}

	s.srv = grpc.NewServer(options...)
	pb.RegisterServerProtoServer(s.srv, s)

	return nil
}

func (s *server) Deinit(ctx context.Context) error {
	s.srv.Stop()
	_ = s.cfg.Logger.Deinit(ctx)
	_ = s.cfg.Scheduler.Deinit(ctx)

	return nil
}

func (s *server) Run(_ context.Context) error {
	lis, _ := net.Listen("tcp", s.cfg.Address)
	return s.srv.Serve(lis)
}

func (s *server) SendServer(ctx context.Context, in *pb.ServerRequest) (*pb.ServerReply, error) {
	var nodes []*common.Node

	if in.GetKind() != Kind {
		return &pb.ServerReply{Error: "invalid kind"}, nil
	}

	nodes, err := s.sendHelper(ctx, in.GetSpec().GetTask(), in.GetSpec().GetNodes())
	if err != nil {
		return &pb.ServerReply{Error: "invalid spec"}, nil
	}

	res := s.cfg.Scheduler.Run(ctx, s.task, nodes)
	_ = s.writeLog(ctx, s.task, nodes, res)

	return &pb.ServerReply{
		Name:  res.Name,
		Error: res.Error,
	}, nil
}

func (s *server) sendHelper(_ context.Context, task *pb.Task, nodes []*pb.Node) ([]*common.Node, error) {
	var buf []*common.Node

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
		NodeSelectors:          task.GetNodeSelectors(),
		RequestedResource:      taskHelper(task),
		ToleratesUnschedulable: task.GetToleratesUnschedulable(),
	}

	for _, item := range nodes {
		node := nodeHelper(item)
		buf = append(buf, &node)
	}

	return buf, nil
}

func (s *server) writeLog(_ context.Context, task *common.Task, nodes []*common.Node, result scheduler.Result) error {
	s.cfg.Logger.Info("server", zap.Any("task", task), zap.Any("nodes", nodes), zap.Any("result", result))
	return nil
}
