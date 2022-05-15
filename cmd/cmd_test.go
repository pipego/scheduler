package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	var err error

	_, err = initConfig("invalid.yml")
	assert.NotEqual(t, nil, err)

	_, err = initConfig("../test/config/invalid.yml")
	assert.NotEqual(t, nil, err)

	_, err = initConfig("../test/config/config.yml")
	assert.Equal(t, nil, err)
}

func TestInitPlugin(t *testing.T) {
	c, err := initConfig("../test/config/config.yml")
	assert.Equal(t, nil, err)

	_, err = initPlugin(c)
	assert.Equal(t, nil, err)
}

func TestInitScheduler(t *testing.T) {
	c, err := initConfig("../test/config/config.yml")
	assert.Equal(t, nil, err)

	_, err = initScheduler(c, nil)
	assert.Equal(t, nil, err)
}

func TestInitServer(t *testing.T) {
	c, err := initConfig("../test/config/config.yml")
	assert.Equal(t, nil, err)

	_, err = initServer(c, nil)
	assert.Equal(t, nil, err)
}
