package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestReaFeaturesConf_RoundTripKeepsFalseAndOmitsAbsent(t *testing.T) {
	t.Parallel()
	off, on := false, true
	cases := []struct {
		name string
		in   ReaFeaturesConf
		want string
	}{
		{"absent", ReaFeaturesConf{}, "{}\n"},
		{"false", ReaFeaturesConf{Enabled: &off}, "rea:\n    enabled: false\n"},
		{"true", ReaFeaturesConf{Enabled: &on}, "rea:\n    enabled: true\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := yaml.Marshal(FeaturesConf{Rea: tc.in})
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(out))
			var back FeaturesConf
			require.NoError(t, yaml.Unmarshal(out, &back))
			assert.Equal(t, tc.in.Decided(), back.Rea.Decided())
			if tc.in.Enabled != nil {
				assert.Equal(t, *tc.in.Enabled, *back.Rea.Enabled)
			}
		})
	}
}

func TestReaFeaturesConf_StrictLoaderRejectsUnknownKeysAndValues(t *testing.T) {
	t.Parallel()
	for name, block := range map[string]string{
		"unknown key": "features:\n  rea:\n    enabled: true\n    version: 6.5.0\n",
		"non-bool":    "features:\n  rea:\n    enabled: sometimes\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, "autopus.yaml"),
				[]byte(strictBaseConfig+block), 0o644))
			_, err := Load(root)
			assert.Error(t, err)
		})
	}
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "autopus.yaml"),
		[]byte(strictBaseConfig+"features:\n  rea:\n    enabled: false\n"), 0o644))
	cfg, err := Load(root)
	require.NoError(t, err)
	require.NotNil(t, cfg.Features.Rea.Enabled)
	assert.False(t, *cfg.Features.Rea.Enabled)
}

func TestDefaultFullConfig_HasNoReaKeyAndKeepAction(t *testing.T) {
	t.Parallel()
	cfg := DefaultFullConfig("p")
	assert.False(t, cfg.Features.Rea.Decided())
	assert.Equal(t, ReaMCPKeep, cfg.Runtime.ReaMCP)
}
