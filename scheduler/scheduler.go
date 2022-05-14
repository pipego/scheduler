package scheduler

import (
	"context"
	"math/rand"

	"github.com/pkg/errors"

	pb "github.com/pipego/scheduler/server/proto"

	"github.com/pipego/scheduler/config"
)

type Scheduler interface {
	Init() error
	Run(*pb.Task, []*pb.Node) Result
}

type Config struct {
	Config config.Config
}

type Result struct {
	Name  string
	Error string
}

type scheduler struct {
	cfg *Config
}

type nodeScore struct {
	name  string
	score int64
}

func New(_ context.Context, cfg *Config) Scheduler {
	return &scheduler{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (s *scheduler) Init() error {
	// TODO: ADD
	return nil
}

func (s *scheduler) Run(*pb.Task, []*pb.Node) Result {
	var list []nodeScore

	// TODO: NumNods == 0

	_, err := s.runFetchPlugins()
	if err != nil {
		return Result{Error: "failed to fetch"}
	}

	_, err = s.runFilterPlugins()
	if err != nil {
		return Result{Error: "failed to filter"}
	}

	_, err = s.runScorePlugins()
	if err != nil {
		return Result{Error: "failed to score"}
	}

	host, err := s.selectHost(list)
	if err != nil {
		return Result{Error: "failed to select"}
	}

	return Result{Name: host}
}

func (s *scheduler) runFetchPlugins() (string, error) {
	// TODO: ADD
	return "", nil
}

func (s *scheduler) runFilterPlugins() (string, error) {
	// TODO: ADD
	return "", nil
}

func (s *scheduler) runScorePlugins() (string, error) {
	// TODO: ADD
	return "", nil
}

// nolint: gosec
func (s *scheduler) selectHost(list []nodeScore) (string, error) {
	if len(list) == 0 {
		return "", errors.New("empty list")
	}

	count := 1
	selected := list[0].name
	max := list[0].score

	for _, item := range list[1:] {
		if item.score > max {
			max = item.score
			selected = item.name
			count = 1
		} else if item.score == max {
			count++
			if rand.Intn(count) == 0 {
				// Replace the candidate with probability of 1/count
				selected = item.name
			}
		}
	}

	return selected, nil
}
