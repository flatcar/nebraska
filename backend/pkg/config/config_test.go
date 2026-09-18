package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

// A value that is not recognised must not fall back to a default: a templated
// deployment can drift to something like "Subscriber" or "replica", and picking
// a role by guessing would serve the wrong one silently.
func TestValidateInstanceMode(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mode    InstanceMode
		wantErr bool
	}{
		{name: "unset", mode: ""},
		{name: "single", mode: InstanceModeSingle},
		{name: "control", mode: InstanceModeControl},
		{name: "edge", mode: InstanceModeEdge},
		{name: "capitalised", mode: "Edge", wantErr: true},
		{name: "upper case", mode: "EDGE", wantErr: true},
		{name: "leading space", mode: " edge", wantErr: true},
		{name: "trailing space", mode: "edge ", wantErr: true},
		{name: "trailing newline", mode: "edge\n", wantErr: true},
		{name: "synonym", mode: "subscriber", wantErr: true},
		{name: "replication vocabulary", mode: "replica", wantErr: true},
		{name: "primary", mode: "primary", wantErr: true},
		{name: "typo", mode: "egde", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := (&Config{AuthMode: "noop", InstanceMode: tc.mode}).Validate()

			if !tc.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid instance-mode")
			assert.Contains(t, err.Error(), strconv.Quote(string(tc.mode)),
				"the value must be quoted so that whitespace in it is visible")
		})
	}
}

func TestValidateSyncerOnEdge(t *testing.T) {
	err := (&Config{AuthMode: "noop", InstanceMode: InstanceModeEdge, EnableSyncer: true}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enable-syncer")

	for _, mode := range []InstanceMode{"", InstanceModeSingle, InstanceModeControl} {
		require.NoError(t, (&Config{AuthMode: "noop", InstanceMode: mode, EnableSyncer: true}).Validate(),
			"the syncer must stay available in %q mode", mode)
	}

	require.NoError(t, (&Config{AuthMode: "noop", InstanceMode: InstanceModeEdge}).Validate())
}

func TestParseInstanceMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		env  string
		want InstanceMode
	}{
		{name: "unset becomes single", want: InstanceModeSingle},
		{name: "taken from the environment", env: string(InstanceModeEdge), want: InstanceModeEdge},
		{name: "taken from the flag", args: []string{"--instance-mode", string(InstanceModeControl)}, want: InstanceModeControl},
		{name: "the flag wins over the environment", args: []string{"--instance-mode", string(InstanceModeControl)}, env: string(InstanceModeEdge), want: InstanceModeControl},
		{name: "an unknown value is left for Validate", env: "Edge", want: "Edge"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withArgs(t, tc.args)
			t.Setenv(instanceModeEnvName, tc.env)

			conf, err := Parse()
			require.NoError(t, err)
			assert.Equal(t, tc.want, conf.InstanceMode)
		})
	}
}

// withArgs replaces the command line for one test, because Parse reads os.Args.
func withArgs(t *testing.T, args []string) {
	t.Helper()

	original := os.Args
	t.Cleanup(func() { os.Args = original })

	os.Args = append([]string{"nebraska"}, args...)
}
