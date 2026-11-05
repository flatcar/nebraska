package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

func TestArchLabel(t *testing.T) {
	tests := []struct {
		name string
		arch int
		want string
	}{
		{"no channel", -1, noChannelArchLabel},
		{"all", int(types.ArchAll), "all"},
		{"amd64", int(types.ArchAMD64), "amd64"},
		{"aarch64", int(types.ArchAArch64), "aarch64"},
		{"x86", int(types.ArchX86), "x86"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, archLabel(tt.arch))
		})
	}
}
