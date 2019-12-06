package gob

import (
	"bytes"
	stdgob "encoding/gob"
	"fmt"

	"github.com/kinvolk/nebraska/pkg/sessions"
	"github.com/kinvolk/nebraska/pkg/sessions/memcache"
)

type copier struct{}

var _ memcache.ValuesCopier = copier{}

// New returns a new ValuesCopier that uses "encoding/gob" to deep
// copy the session values. If using a custom type in session values,
// then register it with gob.Register(), so copying can succeed.
func New() memcache.ValuesCopier {
	return copier{}
}

// Copy is a part of memcache.ValuesCopier interface.
func (copier) Copy(to *sessions.ValuesType, from sessions.ValuesType) error {
	var buf bytes.Buffer
	enc := stdgob.NewEncoder(&buf)
	dec := stdgob.NewDecoder(&buf)
	err := enc.Encode(from)
	if err != nil {
		return fmt.Errorf("could not copy session values, encoding to gob failed: %v", err)
	}
	err = dec.Decode(to)
	if err != nil {
		return fmt.Errorf("could not copy session values, decoding from gob failed: %v", err)
	}
	return nil
}
