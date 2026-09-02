package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestArch(t *testing.T) {
	for _, tt := range []struct {
		arch   types.Arch
		valid  bool
		our    string
		omaha  string
		coreos string
	}{
		{
			arch:   types.ArchAll,
			valid:  true,
			our:    "all",
			omaha:  "",
			coreos: "",
		},
		{
			arch:   types.ArchAMD64,
			valid:  true,
			our:    "amd64",
			omaha:  "x64",
			coreos: "amd64-usr",
		},
		{
			arch:   types.ArchX86,
			valid:  true,
			our:    "x86",
			omaha:  "x86",
			coreos: "",
		},
		{
			arch:   types.ArchAArch64,
			valid:  true,
			our:    "aarch64",
			omaha:  "arm",
			coreos: "arm64-usr",
		},
		{
			arch:   types.Arch(77777),
			valid:  false,
			our:    "Arch(77777)",
			omaha:  "",
			coreos: "",
		},
	} {
		assert.Equal(t, tt.valid, tt.arch.IsValid())
		assert.Equal(t, tt.our, tt.arch.String())
		assert.Equal(t, tt.omaha, tt.arch.OmahaString())
		assert.Equal(t, tt.coreos, tt.arch.CoreosString())
		gotOur, errOur := types.ArchFromString(tt.our)
		gotOmaha, errOmaha := types.ArchFromOmahaString(tt.omaha)
		gotCoreos, errCoreos := types.ArchFromCoreosString(tt.coreos)
		if !tt.valid {
			assert.Equal(t, types.ErrInvalidArch, errOur)
		} else {
			assert.Equal(t, tt.arch, gotOur)
			assert.NoError(t, errOur)
		}
		if !tt.valid || tt.omaha == "" {
			assert.Equal(t, types.ErrInvalidArch, errOmaha)
		} else {
			assert.Equal(t, tt.arch, gotOmaha)
			assert.NoError(t, errOmaha)
		}
		if !tt.valid || tt.coreos == "" {
			assert.Equal(t, types.ErrInvalidArch, errCoreos)
		} else {
			assert.Equal(t, tt.arch, gotCoreos)
			assert.NoError(t, errCoreos)
		}
	}
}
