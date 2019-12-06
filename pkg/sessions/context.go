package sessions

import (
	"context"
)

type storeCtxType int

const (
	sessionCtxKey storeCtxType = iota
)

// SessionFromContext returns the last session stored in the
// context. Returns nil if there is no session.
func SessionFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionCtxKey).(*Session)
	return s
}

// ContextWithSession returns a new context with a session stored in
// it.
func ContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, session)
}
