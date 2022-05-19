package parallelizer

import (
	"context"
	"math"

	"github.com/pipego/scheduler/config"
)

const (
	// DefaultParallelism is the default parallelism used in scheduler.
	DefaultParallelism int = 16
)

type Parallelizer interface {
	Init(p int) error
}

type Config struct {
	Config config.Config
}

type parallelizer struct {
	cfg         *Config
	parallelism int
}

func New(_ context.Context, cfg *Config) Parallelizer {
	return &parallelizer{
		cfg:         cfg,
		parallelism: DefaultParallelism,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *parallelizer) Init(parallelism int) error {
	p.parallelism = parallelism
	return nil
}

func (p *parallelizer) Until(ctx context.Context, pieces int, doWorkPiece DoWorkPieceFunc) {
	ParallelizeUntil(ctx, p.parallelism, pieces, doWorkPiece, WithChunkSize(p.chunkSizeFor(pieces, p.parallelism)))
}

// chunkSizeFor returns a chunk size for the given number of items to use for
// parallel work. The size aims to produce good CPU utilization.
// returns max(1, min(sqrt(num), num/Parallelism))
func (p *parallelizer) chunkSizeFor(num, parallelism int) int {
	s := int(math.Sqrt(float64(num)))

	if r := num/parallelism + 1; s > r {
		s = r
	} else if s < 1 {
		s = 1
	} else {
		// BYPASS
	}

	return s
}
