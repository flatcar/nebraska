package sessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionAndContext(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, SessionFromContext(ctx))
	session := &Session{
		name: "abc",
		id:   "123",
	}
	ctx = ContextWithSession(ctx, session)
	session2 := SessionFromContext(ctx)
	assert.Equal(t, session, session2)
}
