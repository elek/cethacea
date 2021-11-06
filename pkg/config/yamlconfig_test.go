package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadYamlConfig(t *testing.T) {
	k := []struct {
		Name string
		Asd  int
	}{}
	def, err := LoadYamlConfig("test_yaml.yaml", "cubes", &k)
	require.NoError(t, err)
	require.Equal(t, "one", def)
	require.Equal(t, 1, len(k))
	require.Equal(t, "one", k[0].Name)
}
