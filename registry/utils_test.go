package registry

import (
	"github.com/rhine-tech/scene/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetInterfaceName(t *testing.T) {
	assert.Equal(t, "model.DatabaseConfig", getInterfaceName[model.DatabaseConfig]())
}
