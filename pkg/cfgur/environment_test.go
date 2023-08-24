package cfgur

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testStruct struct {
	Username string `cfgur:"username"`
	Value    int    `cfgur:"value"`
	Enabled  bool   `cfgur:"enabled"`
	Val2     int    `cfgur:"val2,default=3"`
}

func TestCommonMarshaller_Unmarshal(t *testing.T) {
	os.Setenv("username", "test")
	os.Setenv("value", "1")
	os.Setenv("enabled", "true")
	var ss testStruct
	err := NewEnvMarshaller().Unmarshal(&ss)
	assert.Nil(t, err)
	assert.Equal(t, "test", ss.Username)
	assert.Equal(t, 1, ss.Value)
	assert.Equal(t, true, ss.Enabled)
	assert.Equal(t, 3, ss.Val2)
}
