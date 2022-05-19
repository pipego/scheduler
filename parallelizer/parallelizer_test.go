package parallelizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkSizeFor(t *testing.T) {
	p := parallelizer{}

	num := 0
	parallelism := 1

	res := p.chunkSizeFor(num, parallelism)
	assert.Equal(t, 1, res)

	num = 1
	parallelism = 1

	res = p.chunkSizeFor(num, parallelism)
	assert.Equal(t, 1, res)

	num = 2
	parallelism = 1

	res = p.chunkSizeFor(num, parallelism)
	assert.Equal(t, 1, res)
}
