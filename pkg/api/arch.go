package api

import (
	"fmt"
)

type Arch uint

const (
	ArchAll Arch = iota
	ArchAMD64
	ArchAArch64
	ArchX86
)

var allSupportedArches = map[Arch]struct{}{
	ArchAll:     {},
	ArchAMD64:   {},
	ArchAArch64: {},
	ArchX86:     {},
}

func sprintfBogusArch(raw uint) string {
	return fmt.Sprintf("Arch(%d)", raw)
}

func emptyBogusArch(raw uint) string {
	return ""
}

const (
	ourArchIdx = iota
	omahaArchIdx
	coreosArchIdx
)

var archStringData = [][3]string{
	// our string, omaha string, coreos string
	{"all", "", ""},
	{"amd64", "x64", "amd64-usr"},
	{"aarch64", "arm", "arm64-usr"},
	{"x86", "x86", ""},
}

func (a Arch) toString(kindIdx int, bogus func(uint) string) string {
	archIdx := int(a)
	if archIdx < len(archStringData) {
		return archStringData[archIdx][kindIdx]
	}
	return bogus(uint(a))
}

func (a Arch) String() string {
	return a.toString(ourArchIdx, sprintfBogusArch)
}

func (a Arch) OmahaString() string {
	return a.toString(omahaArchIdx, emptyBogusArch)
}

func (a Arch) CoreosString() string {
	return a.toString(coreosArchIdx, emptyBogusArch)
}

func (a Arch) IsValid() bool {
	_, ok := allSupportedArches[a]
	return ok
}

func pkgArchFromIdxString(s string, idx int) (Arch, error) {
	if s == "" {
		return ArchAll, ErrInvalidArch
	}
	for i := 0; i < len(archStringData); i++ {
		if s == archStringData[i][idx] {
			return Arch(i), nil
		}
	}
	return ArchAll, ErrInvalidArch
}

func ArchFromString(s string) (Arch, error) {
	return pkgArchFromIdxString(s, ourArchIdx)
}

func ArchFromOmahaString(s string) (Arch, error) {
	return pkgArchFromIdxString(s, omahaArchIdx)
}

func ArchFromCoreosString(s string) (Arch, error) {
	return pkgArchFromIdxString(s, coreosArchIdx)
}
