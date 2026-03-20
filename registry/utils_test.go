package registry

import (
	"testing"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/stretchr/testify/assert"
)

func TestGetInterfaceName(t *testing.T) {
	assert.Equal(t, "datasource.DatabaseConfig", getInterfaceName[datasource.DatabaseConfig]())
}
