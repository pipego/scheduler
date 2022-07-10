package parallelizer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestParallelizeUntil(t *testing.T) {
	pieces := []string{
		"piece1",
		"piece2",
		"piece3",
	}

	defer goleak.VerifyNone(t)

	ParallelizeUntil(context.Background(), DefaultParallelism, len(pieces), func(index int) {
		fmt.Println(pieces[index])
	})
}

func TestCeilDiv(t *testing.T) {
	a := 0
	b := 1

	res := ceilDiv(a, b)
	assert.Equal(t, 0, res)

	a = 1
	b = 1

	res = ceilDiv(a, b)
	assert.Equal(t, 1, res)

	a = 1
	b = 2

	res = ceilDiv(a, b)
	assert.Equal(t, 1, res)

	a = 2
	b = 3

	res = ceilDiv(a, b)
	assert.Equal(t, 1, res)
}
