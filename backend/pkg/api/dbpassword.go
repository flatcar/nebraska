package api

import (
	"context"
	"sync"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
)

// DBPasswordProvider supplies the password used to authenticate a database
// connection. Nebraska calls it before opening every physical connection, so an
// implementation is free to return a credential that expires, such as an OAuth
// token, without Nebraska having to know how it is obtained.
//
// The user the connection is being made as is passed so that a single provider
// can serve connections that authenticate as different roles.
//
// If connect_timeout is set in NEBRASKA_DB_URL, the context carries it as its
// deadline, and an implementation is expected to honour it.
type DBPasswordProvider interface {
	DBPassword(ctx context.Context, user string) (string, error)
}

var (
	dbPasswordMu       sync.RWMutex
	dbPasswordProvider DBPasswordProvider
)

// SetDBPasswordProvider installs the provider New authenticates with. Call it
// before New.
func SetDBPasswordProvider(provider DBPasswordProvider) {
	dbPasswordMu.Lock()
	defer dbPasswordMu.Unlock()

	dbPasswordProvider = provider
}

// dbPasswordFunc adapts the installed provider to what dbconn expects, and
// returns nil when there is none.
func dbPasswordFunc() dbconn.PasswordFunc {
	dbPasswordMu.RLock()
	defer dbPasswordMu.RUnlock()

	if dbPasswordProvider == nil {
		return nil
	}

	return dbPasswordProvider.DBPassword
}
