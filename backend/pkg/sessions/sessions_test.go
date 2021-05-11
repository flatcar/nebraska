package sessions

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

func newMockSessionFast() (*MockCache, *MockCodec, *Session) {
	cache := NewMockCache()
	codec := NewMockCodec()
	codec.AddIDValueMapping("id1", "val1")
	return cache, codec, newMockSessionExt("test", cache, codec).Session()
}

func newMockSessionExt(name string, cache *MockCache, codec *MockCodec) SessionExt {
	builder := sessionBuilder{
		codec: codec,
	}
	return builder.NewSession(name, cache)
}
