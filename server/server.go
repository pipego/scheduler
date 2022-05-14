package server

import (
	"context"
	"math"
	"net"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	pb "github.com/pipego/scheduler/server/proto"

	"github.com/pipego/scheduler/config"
	"github.com/pipego/scheduler/plugin"
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
	Plugin    plugin.Plugin
	Scheduler scheduler.Scheduler
}

type server struct {
	cfg *Config
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
	if err := s.cfg.Plugin.Init(); err != nil {
		return errors.Wrap(err, "failed to init plugin")
	}

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

	res := s.cfg.Scheduler.Run(in.GetSpec().GetTask(), in.GetSpec().GetNodes())

	return &pb.ServerReply{
		Name:  res.Name,
		Error: res.Error,
	}, nil
}
