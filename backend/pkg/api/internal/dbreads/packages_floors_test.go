package dbreads

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

// TestSelectFloorsInRange covers the floor version range, ordering, and limit logic,
// including pre-release and build metadata precedence.
func TestSelectFloorsInRange(t *testing.T) {
	mk := func(vs ...string) []*types.Package {
		ps := make([]*types.Package, len(vs))
		for i, v := range vs {
			ps[i] = &types.Package{Version: v}
		}
		return ps
	}
	vers := func(ps []*types.Package) []string {
		out := make([]string, 0, len(ps))
		for _, p := range ps {
			out = append(out, p.Version)
		}
		return out
	}

	tests := []struct {
		name             string
		floors           []*types.Package
		instance, target string
		limit            int
		want             []string
	}{
		{"release floor not skipped over pre-release instance", mk("2000.0.0"), "2000.0.0-rc1", "3000.0.0", 5, []string{"2000.0.0"}},
		{"same-core floors ordered, none dropped", mk("2000.0.0", "2000.0.0-beta", "2000.0.0-alpha"), "1000.0.0", "3000.0.0", 5, []string{"2000.0.0-alpha", "2000.0.0-beta", "2000.0.0"}},
		{"lower pre-release below release instance excluded", mk("2000.0.0-rc1"), "2000.0.0", "3000.0.0", 5, []string{}},
		{"build metadata treated as the release", mk("2000.0.0+build"), "1999.0.0", "3000.0.0", 5, []string{"2000.0.0+build"}},
		{"shuffled input sorted and limited", mk("3000.0.0", "1000.0.0", "2000.0.0"), "500.0.0", "4000.0.0", 2, []string{"1000.0.0", "2000.0.0"}},
		{"target inclusive, above-target excluded", mk("2000.0.0", "4000.0.0"), "1000.0.0", "3000.0.0", 5, []string{"2000.0.0"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := selectFloorsInRange(tt.floors, tt.instance, tt.target, tt.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.want, vers(res))
		})
	}
}
